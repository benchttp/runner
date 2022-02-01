package requester

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/dispatcher"
)

// Requester executes the benchmark. It wraps http.Client.
type Requester struct {
	records chan Record // Records provides read access to the results of Requester.Run.

	config config.Config
	client http.Client
	tracer *tracer
}

// New returns a Requester configured with specified Options.
func New(cfg config.Config) *Requester {
	tracer := newTracer()
	return &Requester{
		records: make(chan Record, cfg.RunnerOptions.Requests),
		config:  cfg,
		tracer:  tracer,
		client: http.Client{
			// Timeout includes connection time, any redirects, and reading
			// the response body.
			// We may want exclude reading the response body in our benchmark tool.
			Timeout: cfg.Request.Timeout,

			// tracer keeps track of all events of the current request.
			Transport: tracer,
		},
	}
}

// Run starts the benchmark test and pipelines the results inside a Report.
// Returns the Report when the test ended and all results have been collected.
func (r *Requester) Run() (Report, error) {
	req, err := r.config.HTTPRequest()
	if err != nil {
		return Report{}, err
	}

	if err := r.ping(req); err != nil {
		return Report{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.RunnerOptions.GlobalTimeout)
	errCh := make(chan error)

	go func() {
		defer cancel()
		defer close(r.records)

		errCh <- dispatcher.
			New(r.config.RunnerOptions.Concurrency).
			Do(ctx, r.config.RunnerOptions.Requests, r.record(req))
	}()

	if err := <-errCh; err != nil {
		return Report{}, err
	}

	return r.collect(), nil
}

func (r *Requester) ping(req *http.Request) error {
	resp, err := r.client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	return err
}

// Record is the summary of a HTTP response. If Record.Error is non-nil,
// the HTTP call failed anywhere from making the request to decoding the
// response body, invalidating the entire response, as it is not a remote
// server error.
type Record struct {
	Time   time.Duration `json:"time"`
	Code   int           `json:"code"`
	Bytes  int           `json:"bytes"`
	Error  error         `json:"error,omitempty"`
	Events []Event       `json:"events"`
}

func (r *Requester) record(req *http.Request) func() {
	return func() {
		sent := time.Now()

		resp, err := r.client.Do(req)
		if err != nil {
			r.records <- Record{Error: err}
			return
		}

		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			r.records <- Record{Error: err}
			return
		}

		r.records <- Record{
			Code:   resp.StatusCode,
			Time:   time.Since(sent),
			Bytes:  len(body),
			Events: r.tracer.events,
		}
	}
}
