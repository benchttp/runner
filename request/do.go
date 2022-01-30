package request

import (
	"context"
	"net/http"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/semimpl"
)

func doRequest(url string, timeout time.Duration) Record {
	client := http.Client{
		// Timeout includes connection time, any redirects, and reading the response body.
		// We may want exclude reading the response body in our benchmark tool.
		Timeout: timeout,
	}

	start := time.Now()

	resp, err := client.Get(url) //nolint:bodyclose
	end := time.Since(start)
	if err != nil {
		return Record{Error: err}
	}

	return newRecord(resp, end)
}

// Do launches a goroutine to ping url as soon as a thread is
// available and collects the results as they come in.
// The value of concurrency limits the number of concurrent threads.
// Once all requests have been made or on done signal from ctx,
// waits for goroutines to end and returns the collected records.
func Do(ctx context.Context, cfg config.Config) <-chan Record {
	var (
		uri        = cfg.Request.URL.String()
		numWorker  = cfg.RunnerOptions.Concurrency
		numRequest = cfg.RunnerOptions.Requests
		reqTimeout = cfg.Request.Timeout
	)

	ch := make(chan Record, numRequest)

	go func() {
		defer close(ch)

		semimpl.Do(ctx, numWorker, numRequest, func() {
			ch <- doRequest(uri, reqTimeout)
		})
	}()

	return ch
}

// Collect collects records from ch and returns how many were collected.
func Collect(ch <-chan Record) int {
	var length int

	for range ch {
		length++
	}
	return length
}
