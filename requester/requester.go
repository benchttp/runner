package requester

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/benchttp/runner/dispatcher"
)

const (
	defaultRecordsCap = 1000
)

type Config struct {
	Requests       int
	Concurrency    int
	Interval       time.Duration
	RequestTimeout time.Duration
	GlobalTimeout  time.Duration
}

// Requester executes the benchmark. It wraps http.Client.
type Requester struct {
	records []Record
	numErr  int
	runErr  error
	start   time.Time
	done    bool

	config       Config
	newTransport func() http.RoundTripper

	mu sync.RWMutex
}

// New returns a Requester initialized with cfg. cfg is assumed valid:
// it is the caller's responsibility to ensure cfg is valid using
// cfg.Validate.
func New(cfg Config) *Requester {
	recordsCap := cfg.Requests
	if recordsCap < 1 {
		recordsCap = defaultRecordsCap
	}

	return &Requester{
		records: make([]Record, 0, recordsCap),
		config:  cfg,
		newTransport: func() http.RoundTripper {
			return newTracer()
		},
	}
}

// Run starts the benchmark test and pipelines the results inside a Report.
// Returns the Report when the test ended and all results have been collected.
func (r *Requester) Run(req *http.Request) (Report, error) {
	if err := r.ping(req); err != nil {
		return Report{}, fmt.Errorf("%w: %s", ErrConnection, err)
	}

	var (
		numWorker   = r.config.Concurrency
		maxIter     = r.config.Requests
		timeout     = r.config.GlobalTimeout
		interval    = r.config.Interval
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	)

	defer cancel()

	r.start = time.Now()
	go r.refreshState()

	switch err := dispatcher.New(numWorker).Do(ctx, maxIter, r.record(req, interval)); err {
	case nil, context.Canceled, context.DeadlineExceeded:
		r.end(err)
	default:
		return Report{}, err
	}

	return makeReport(r.records, r.numErr), nil
}

func (r *Requester) ping(req *http.Request) error {
	resp, err := newClient(r.newTransport(), r.config.RequestTimeout).Do(req)
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

func (r *Requester) record(req *http.Request, interval time.Duration) func() {
	return func() {
		client := newClient(r.newTransport(), r.config.RequestTimeout)
		// It is necessary to clone the request because one request
		// with a non-nil body cannot be used in concurrent threads.
		newReq := cloneRequest(req)

		sent := time.Now()

		resp, err := client.Do(newReq)
		if err != nil {
			r.appendRecord(Record{Error: err})
			return
		}

		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			r.appendRecord(Record{Error: err})
			return
		}

		duration := time.Since(sent)

		r.appendRecord(Record{
			Code:   resp.StatusCode,
			Time:   duration,
			Bytes:  len(body),
			Events: toTracer(client.Transport).events,
		})

		r.printState()
		time.Sleep(interval)
	}
}

func (r *Requester) appendRecord(rec Record) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
	if rec.Error != nil {
		r.numErr++
	}
}

func (r *Requester) refreshState() {
	ticker := time.NewTicker(time.Second)
	tick := ticker.C
	for {
		if r.done {
			ticker.Stop()
			break
		}
		r.printState()
		<-tick
	}
}

func (r *Requester) end(runErr error) {
	r.mu.Lock()
	r.runErr = runErr
	r.done = true
	r.mu.Unlock()
	r.printState()
}

func (r *Requester) printState() {
	fmt.Print(r.state())
}

// newClient returns a new http.Client with the given transport and timeout.
func newClient(transport http.RoundTripper, timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// cloneRequest fully clones a *http.Request by also cloning the body
// via Request.GetBody.
func cloneRequest(req *http.Request) *http.Request {
	reqClone := req.Clone(req.Context())
	if req.Body != nil {
		// err is always nil (https://golang.org/src/net/http/request.go#L889)
		reqClone.Body, _ = req.GetBody()
	}
	return reqClone
}

// newClient returns a new http.Client with the given transport and timeout.
func toTracer(transport http.RoundTripper) *tracer {
	reqtracer, ok := transport.(*tracer)
	if !ok || reqtracer == nil {
		return &tracer{}
	}
	return reqtracer
}
