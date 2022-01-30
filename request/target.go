package request

import (
	"bytes"
	"net/http"
	"net/url"
)

type Target struct {
	Method string   `json:"method"`
	URL    *url.URL `json:"url"`
	Body   []byte   `json:"body,omitempty"` // Body is not used for now.
}

func (t *Target) Request() (*http.Request, error) {
	req, err := http.NewRequest(t.Method, t.URL.String(), bytes.NewReader(t.Body))
	if err != nil {
		return nil, err
	}
	return req, nil
}
