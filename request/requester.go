package request

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/semimpl"
)

type Requester struct {
	Records chan Record

	client      http.Client
	concurrency int
	requests    int
}

func New(cfg config.RunnerOptions) *Requester {
	r := &Requester{
		Records:     make(chan Record, cfg.Requests),
		concurrency: cfg.Concurrency,
		requests:    cfg.Requests,
	}

	r.client = http.Client{
		// Timeout includes connection time, any redirects, and reading the response body.
		// We may want exclude reading the response body in our benchmark tool.
		Timeout: cfg.GlobalTimeout, // FIXME bad config struct
	}

	return r
}

func (r *Requester) Run(ctx context.Context, cfg config.Request) {
	t := Target{
		URL:    cfg.URL,
		Method: cfg.Method,
	}

	go func() {
		defer close(r.Records)
		semimpl.Do(ctx, r.concurrency, r.requests, func() {
			r.record(t)
		})
	}()
}

// Record is a summary of an http call.
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
