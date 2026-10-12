package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const boundTestInitialize = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}`

// boundTestPing is a ping request padded to more than size bytes.
func boundTestPing(size int64) string {
	return `{"jsonrpc":"2.0","id":2,"method":"ping","params":{"pad":"` + strings.Repeat("x", int(size)) + `"}}`
}

// postRawBody posts body as a JSON-RPC message and returns the status code.
func postRawBody(t *testing.T, url, body string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	return resp.StatusCode
}

func TestStreamableHTTP_MaxRequestBodyBytes(t *testing.T) {
	tests := []struct {
		name    string
		opts    []StreamableHTTPOption
		body    string
		refused bool
	}{
		{name: "beyond a configured bound", opts: []StreamableHTTPOption{WithMaxRequestBodyBytes(1024)}, body: boundTestPing(1024), refused: true},
		{name: "within a configured bound", opts: []StreamableHTTPOption{WithMaxRequestBodyBytes(1024)}, body: boundTestInitialize},
		{name: "beyond the default bound", body: boundTestPing(DefaultMaxRequestBodyBytes), refused: true},
		{name: "unbounded with zero", opts: []StreamableHTTPOption{WithMaxRequestBodyBytes(0)}, body: boundTestPing(DefaultMaxRequestBodyBytes)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(NewStreamableHTTPServer(NewMCPServer("test", "1.0.0"), tt.opts...))
			defer ts.Close()

			status := postRawBody(t, ts.URL+"/mcp", tt.body)
			if tt.refused {
				assert.Equal(t, http.StatusRequestEntityTooLarge, status)
				assert.Equal(t, http.StatusOK, postRawBody(t, ts.URL+"/mcp", boundTestInitialize), "the endpoint keeps answering")
			} else {
				assert.NotEqual(t, http.StatusRequestEntityTooLarge, status)
			}
		})
	}
}

func TestSSE_MaxRequestBodyBytes(t *testing.T) {
	ts := NewTestServer(NewMCPServer("test", "1.0.0"), WithSSEMaxRequestBodyBytes(1024))
	defer ts.Close()

	sseResp, err := http.Get(fmt.Sprintf("%s/sse", ts.URL))
	require.NoError(t, err)
	defer sseResp.Body.Close()
	endpointEvent, err := readSSEEvent(sseResp)
	require.NoError(t, err)
	messageURL := strings.TrimSpace(strings.Split(strings.Split(endpointEvent, "data: ")[1], "\n")[0])

	assert.Equal(t, http.StatusRequestEntityTooLarge, postRawBody(t, messageURL, boundTestPing(1024)))
	assert.Equal(t, http.StatusAccepted, postRawBody(t, messageURL, boundTestInitialize), "the endpoint keeps answering")
}
