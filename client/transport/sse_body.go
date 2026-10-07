package transport

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"sync"
	"time"
)

// sseCloseGrace is how long closing an SSE body waits for an in-flight Read
// before it cancels the request. A server ends a finished stream right after
// its last event, so the reader usually sees EOF well within it, and the
// connection stays reusable.
const sseCloseGrace = 100 * time.Millisecond

var errSSEBodyClosed = errors.New("transport: read on closed SSE response body")

// wrapResponseBody makes resp's body safe to close while a Read on it is in
// flight, and releases cancel, the request's own context, when it is closed.
//
// readSSE closes an SSE body to interrupt a reader blocked in Read, and the
// caller closes it again when it returns. On Go's HTTP/1 transport, a Close
// while a Read is in flight starts a short drain of the body so that the
// connection can be reused. When the stream ends inside that drain, the
// blocked reader's EOF can wait for a signal the connection's read loop has
// already given, holding the body's lock while the connection goes back to
// the pool: the reader, the second Close and the next response on that
// connection wait until the idle connection is closed (90s by default).
// sseBody closes the transport's body once, only after its reader returned.
func wrapResponseBody(resp *http.Response, cancel context.CancelFunc) {
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err == nil && mediaType == "text/event-stream" {
		resp.Body = &sseBody{body: resp.Body, cancel: cancel}
		return
	}
	resp.Body = &cancelOnCloseBody{ReadCloser: resp.Body, cancel: cancel}
}

// cancelOnCloseBody releases the request's context when its body is closed.
type cancelOnCloseBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelOnCloseBody) Close() error {
	err := b.ReadCloser.Close()
	b.cancel()
	return err
}

// sseBody never closes the transport's body while a Read on it is in flight.
// A Close during a Read returns at once and leaves the body to the reader: a
// stream that ends within sseCloseGrace ends cleanly; otherwise the request is
// cancelled, which makes the transport discard the connection and releases
// the reader. Only then is the transport's body closed.
type sseBody struct {
	body   io.ReadCloser
	cancel context.CancelFunc

	mu      sync.Mutex
	closed  bool
	reading chan struct{} // closed when the in-flight Read returns; nil when none is
}

func (b *sseBody) Read(p []byte) (int, error) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return 0, errSSEBodyClosed
	}
	reading := make(chan struct{})
	b.reading = reading
	b.mu.Unlock()

	n, err := b.body.Read(p)

	b.mu.Lock()
	b.reading = nil
	b.mu.Unlock()
	close(reading)
	return n, err
}

// Close never blocks on an in-flight Read: it hands the transport's body to a
// goroutine that closes it once the Read has returned.
func (b *sseBody) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	reading := b.reading
	b.mu.Unlock()

	if reading == nil {
		err := b.body.Close()
		b.cancel()
		return err
	}
	go func() {
		timer := time.NewTimer(sseCloseGrace)
		defer timer.Stop()
		select {
		case <-reading:
		case <-timer.C:
			b.cancel()
			<-reading
		}
		_ = b.body.Close()
		b.cancel()
	}()
	return nil
}
