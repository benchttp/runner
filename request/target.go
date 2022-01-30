package request

import (
	"bytes"
	"net/http"
	"net/url"
)

// Target defines the target of the benchmark test.
// It is a subset of http.Request.
type Target struct {
	Method string   `json:"method"`
	URL    *url.URL `json:"url"`
	Body   []byte   `json:"body,omitempty"` // Body is not used for now.
}

// Request returns a *http.Request created from Target. Returns any non-nil
// error that occurred.
func (t *Target) Request() (*http.Request, error) {
	req, err := http.NewRequest(t.Method, t.URL.String(), bytes.NewReader(t.Body))
	if err != nil {
		return nil, err
	}
	return req, nil
}
