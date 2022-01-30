package requester

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/benchttp/runner/semimpl"
)

// Options are a set of properties that define Requester behavior.
type Options struct {
	Requests    int
	Concurrency int
	Duration    time.Duration
	Timeout     time.Duration
}

// Requester executes the benchmark. It wraps http.Client.
type Requester struct {
	Records chan Record // Records provides read access to the results of Requester.Run.

	client      http.Client
	concurrency int
	requests    int
}

// New returns a Requester configured with specified Options.
func New(o Options) *Requester {
	r := &Requester{
		Records:     make(chan Record, o.Requests),
		concurrency: o.Concurrency,
		requests:    o.Requests,
	}

	r.client = http.Client{
		// Timeout includes connection time, any redirects, and reading the response body.
		// We may want exclude reading the response body in our benchmark tool.
		Timeout: o.Timeout,
	}

	return r
}

// Run launches the benchmark test. The test runs inside a goroutine managing
// its own concurrent workers. Run does not block, the results of the test can
// be pipelined from Requester.Records for some other usage.
func (r *Requester) Run(ctx context.Context, t Target) {
	go func() {
		defer close(r.Records)
		semimpl.Do(ctx, r.concurrency, r.requests, func() {
			r.record(t)
		})
	}()
}

// Record is the summary of a HTTP response. If Record.Error is non-nil,
// the HTTP call failed anywhere from making the request to decoding the
// response body, invalidating the entire response, as it is not a remote
// server error.
type Record struct {
	Time  time.Duration `json:"time"`
	Code  int           `json:"code"`
	Bytes int           `json:"bytes"`
	Error error         `json:"error"`
}

func (r *Requester) record(t Target) {
	req, err := t.Request()
	if err != nil {
		r.Records <- Record{Error: err}
		return
	}

	sent := time.Now()

	resp, err := r.client.Do(req)
	if err != nil {
		r.Records <- Record{Error: err}
		return
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		r.Records <- Record{Error: err}
		return
	}

	r.Records <- Record{
		Code:  resp.StatusCode,
		Time:  time.Since(sent),
		Bytes: len(body),
	}
}
