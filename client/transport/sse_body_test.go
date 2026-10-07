package transport

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sseBodyWait bounds every blocking step of these tests: far above
// sseCloseGrace and the HTTP/1 transport's 50ms post-close drain, far below
// the 90s idle-connection timeout a held call waits for.
const sseBodyWait = 5 * time.Second

// readSignalConn closes reading once a Read starts after armed is set.
type readSignalConn struct {
	net.Conn
	armed   *atomic.Bool
	once    *sync.Once
	reading chan struct{}
}

func (c readSignalConn) Read(p []byte) (int, error) {
	if c.armed.Load() {
		c.once.Do(func() { close(c.reading) })
	}
	return c.Conn.Read(p)
}

// wrapTransport applies wrapResponseBody, as sendHTTP does, to the responses
// of next.
type wrapTransport struct {
	next http.RoundTripper
}

func (t wrapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(req.Context())
	resp, err := t.next.RoundTrip(req.WithContext(ctx))
	if err != nil {
		cancel()
		return nil, err
	}
	wrapResponseBody(resp, cancel)
	return resp, nil
}

// sseStreamFixture is an SSE server that sends one event, then holds the
// stream open until end is closed, and a keep-alive HTTP/1 client to it whose
// connections report when a reader blocks in Read after armed is set.
type sseStreamFixture struct {
	url     string
	end     chan struct{}
	client  *http.Client
	armed   atomic.Bool
	reading chan struct{}
	dials   atomic.Int32
}

func newSSEStreamFixture(t *testing.T) *sseStreamFixture {
	t.Helper()
	f := &sseStreamFixture{end: make(chan struct{}), reading: make(chan struct{})}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, "{}")
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "event: message\ndata: {}\n\n")
		w.(http.Flusher).Flush()
		select {
		case <-f.end:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(srv.Close)
	f.url = srv.URL

	var once sync.Once
	base := http.DefaultTransport.(*http.Transport).Clone()
	dial := base.DialContext
	base.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := dial(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		f.dials.Add(1)
		return readSignalConn{Conn: conn, armed: &f.armed, once: &once, reading: f.reading}, nil
	}
	t.Cleanup(base.CloseIdleConnections)
	f.client = &http.Client{Transport: wrapTransport{next: base}}
	return f
}

// openStream starts the stream and a reader on it that reads like readSSE:
// it takes the event, then blocks in Read for the next one. It returns the
// body and a channel closed when the reader returns.
func (f *sseStreamFixture) openStream(t *testing.T) (io.ReadCloser, chan struct{}) {
	t.Helper()
	resp, err := f.client.Get(f.url)
	require.NoError(t, err)
	gotEvent := make(chan struct{})
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		br := bufio.NewReader(resp.Body)
		for line := ""; line != "\n"; {
			var err error
			if line, err = br.ReadString('\n'); err != nil {
				return
			}
		}
		f.armed.Store(true)
		close(gotEvent)
		for {
			if _, err := br.ReadString('\n'); err != nil {
				return
			}
		}
	}()
	select {
	case <-gotEvent:
	case <-readerDone:
		t.Fatal("the stream ended before its event")
	}
	<-f.reading
	return resp.Body, readerDone
}

func waitForSSEBody(t *testing.T, done <-chan struct{}, failure string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(sseBodyWait):
		t.Fatal(failure)
	}
}

// closeInBackground closes body and reports when Close returns.
func closeInBackground(body io.Closer) chan struct{} {
	closed := make(chan struct{})
	go func() {
		_ = body.Close()
		close(closed)
	}()
	return closed
}

// TestSSEBody_StreamEndingInTheDrainReleasesTheCall reproduces the frame that
// holds a call for the idle-connection timeout: readSSE closes the body while
// its reader is blocked in Read, and the stream ends right after, inside the
// transport's post-close drain. The reader, SendRequest's second Close and
// the next request on the client must all return at once.
func TestSSEBody_StreamEndingInTheDrainReleasesTheCall(t *testing.T) {
	f := newSSEStreamFixture(t)
	body, readerDone := f.openStream(t)

	_ = body.Close()
	close(f.end)

	waitForSSEBody(t, readerDone, "the reader of a closed stream is still blocked after the stream ended")
	waitForSSEBody(t, closeInBackground(body), "closing the body again blocks: the call would not return before the idle connection is closed")

	ctx, cancel := context.WithTimeout(t.Context(), sseBodyWait)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url+"/json", nil)
	require.NoError(t, err)
	resp, err := f.client.Do(req)
	require.NoError(t, err, "the next request on the client is held")
	_, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "the next response on the client is held")
	_ = resp.Body.Close()
}

// TestSSEBody_OpenStreamIsCancelledAfterTheGrace covers a closed stream the
// server keeps open: Close returns at once, and the request is cancelled,
// which releases the reader.
func TestSSEBody_OpenStreamIsCancelledAfterTheGrace(t *testing.T) {
	f := newSSEStreamFixture(t)
	defer close(f.end)
	body, readerDone := f.openStream(t)

	closed := closeInBackground(body)
	select {
	case <-closed:
	case <-readerDone:
		t.Fatal("the reader returned before Close did")
	case <-time.After(sseBodyWait):
		t.Fatal("Close blocks on the in-flight Read")
	}
	waitForSSEBody(t, readerDone, "the reader of a closed open stream is never released")
}

// TestSSEBody_CloseWithoutAReader covers the plain Close of a body nobody
// reads: it closes the transport's body before returning, a later Read fails
// and a second Close is a no-op.
func TestSSEBody_CloseWithoutAReader(t *testing.T) {
	f := newSSEStreamFixture(t)
	defer close(f.end)
	resp, err := f.client.Get(f.url)
	require.NoError(t, err)

	require.NoError(t, resp.Body.Close())
	_, err = resp.Body.Read(make([]byte, 1))
	assert.ErrorIs(t, err, errSSEBodyClosed)
	assert.NoError(t, resp.Body.Close())
}

// TestSSEBody_OtherResponsesKeepTheirConnection checks that a response read
// to its end leaves its connection to the next request.
func TestSSEBody_OtherResponsesKeepTheirConnection(t *testing.T) {
	f := newSSEStreamFixture(t)
	for range 2 {
		resp, err := f.client.Get(f.url + "/json")
		require.NoError(t, err)
		_, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		_ = resp.Body.Close()
	}
	assert.Equal(t, int32(1), f.dials.Load(), "two requests dialled more than one connection")
}
