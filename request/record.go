package request

import (
	"io"
	"net/http"
	"time"
)

// Record is a summary of an http call.
type Record struct {
	Cost  time.Duration
	Code  int
	Bytes int
	Error error
}

// newRecord returns a Record that summarizes the given http response,
// attaching the duration and a non-nil error if any occurs
// in the reading process.
func newRecord(resp *http.Response, t time.Duration) Record {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	r := Record{
		Code:  resp.StatusCode,
		Cost:  t,
		Bytes: len(body),
	}

	if err != nil {
		r.Error = err
	}

	return r
}
