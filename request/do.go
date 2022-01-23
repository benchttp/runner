package request

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/benchttp/runner/record"
)

func doRequest(url string, timeout time.Duration) record.Record {
	client := http.Client{
		// Timeout includes connection time, any redirects, and reading the response body.
		// We may want exclude reading the response body in our benchmark tool.
		Timeout: timeout,
	}

	start := time.Now()

	resp, err := client.Get(url)
	end := time.Since(start)
	if err != nil {
		return record.Record{Error: err}
	}

	return record.New(resp, end)
}

// Do launches a goroutine to ping url as soon as a thread is
// available and collects the results as they come in.
// The value of concurrency limits the number of concurrent threads.
// Once all requests have been made or on done signal from ctx,
// waits for goroutines to end and returns the collected records.
func Do(ctx context.Context, requests, concurrency int, url string, timeout time.Duration) []record.Record {
	sem := semaphore.NewWeighted(int64(concurrency))
	rec := record.NewSafeSlice(requests)
	wg := sync.WaitGroup{}

	for i := 0; i < requests; i++ {
		fmt.Println(i) // TODO: delete temporary print
		wg.Add(1)

		if err := sem.Acquire(ctx, 1); err != nil {
			handleContextError(err)
			wg.Done()
			break
		}

		go func() {
			defer func() {
				sem.Release(1)
				wg.Done()
			}()
			rec.Append(doRequest(url, timeout))
		}()
	}

	wg.Wait()
	return rec.Slice()
}

// DoUntil launches a goroutine to ping url as soon as a thread is
// available and collects the results as they come in.
// The value of concurrency limits the number of concurrent threads.
// On done signal from ctx, waits for goroutines to end and returns
// the collected records.
func DoUntil(ctx context.Context, concurrency int, url string, timeout time.Duration) []record.Record {
	// sem is a semaphore to constrain access to at most n concurrent threads.
	sem := semaphore.NewWeighted(int64(concurrency))
	rec := record.NewSafeSlice(0)
	wg := sync.WaitGroup{}

	for i := 0; ; i++ { // TODO: back to "for"
		fmt.Println(i) // TODO: delete temporary print
		wg.Add(1)

		if err := sem.Acquire(ctx, 1); err != nil {
			handleContextError(err)
			wg.Done()
			break
		}

		go func() {
			defer func() {
				sem.Release(1)
				wg.Done()
			}()
			rec.Append(doRequest(url, timeout))
		}()
	}

	wg.Wait()
	return rec.Slice()
}

func handleContextError(err error) {
	switch {
	case err == nil:
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Println("timeout") // TODO: remove print
	case errors.Is(err, context.Canceled):
		fmt.Println("cancel") // TODO: remove print
	default:
		log.Fatal(err)
	}
}
