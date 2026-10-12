package server

import (
	"errors"
	"io"
	"net/http"
)

// limitRequestBody returns r's body bounded at limit bytes; a read beyond
// it fails with *http.MaxBytesError. Zero or a negative limit leaves the
// body unbounded.
func limitRequestBody(w http.ResponseWriter, r *http.Request, limit int64) io.Reader {
	if limit <= 0 {
		return r.Body
	}
	return http.MaxBytesReader(w, r.Body, limit)
}

// isRequestBodyTooLarge reports whether err is a read beyond a body's bound.
func isRequestBodyTooLarge(err error) bool {
	var tooLarge *http.MaxBytesError
	return errors.As(err, &tooLarge)
}
