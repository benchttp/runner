package request

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

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

// acquire acquires the semaphore with a weight of 1, blocking until
// the ressource is free, and adds 1 to the WaitGroup counter.
func acquire(sem chan<- int, wg *sync.WaitGroup) {
	sem <- 1
	wg.Add(1)
}

// release releases the semaphore with a weight of 1, freeing the ressource
// for other actors, and decrements the WaitGroup counter by 1.
func release(sem <-chan int, wg *sync.WaitGroup) {
	<-sem
	wg.Done()
}

// Do launches a goroutine to ping url as soon as a thread is
// available and collects the results as they come in.
// The value of concurrency limits the number of concurrent threads.
// Once all requests have been made or on done signal from ctx,
// waits for goroutines to end and returns the collected records.
func Do(ctx context.Context, requests, concurrency int, url string, timeout time.Duration) []record.Record {
	// sem is a semaphore to constrain access to at most n concurrent threads.
	sem := make(chan int, concurrency)
	rec := record.NewSafeSlice(requests)

	var wg sync.WaitGroup

	func() {
		for i := 0; i < requests; i++ {
			fmt.Println(i) // TODO: delete temporary print
			select {
			case <-ctx.Done():
				return
			default:
			}
			acquire(sem, &wg)
			go func() {
				defer release(sem, &wg)
				rec.Append(doRequest(url, timeout))
			}()
		}
	}()

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
	sem := make(chan int, concurrency)
	rec := record.NewSafeSlice(0)

	var wg sync.WaitGroup

	func() {
		for i := 0; ; i++ { // TODO: back to "for"
			select {
			case <-ctx.Done():
				return
			default:
			}
			acquire(sem, &wg)
			fmt.Println(i) // TODO: delete temporary print
			go func() {
				defer release(sem, &wg)
				rec.Append(doRequest(url, timeout))
			}()
		}
	}()

	wg.Wait()
	return rec.Slice()
}
