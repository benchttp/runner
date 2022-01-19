package sem

import (
	"sync"
	"time"

	"github.com/benchttp/runner/request"
)

// RunFor launches a goroutine to ping url as soon as a thread is
// available and sends the results through a pipeline as they come in.
// Returns a channel to pipeline the records.
// The value of concurrency limits the number of concurrent threads.
// Once all requests have been made or on quit signal, waits for
// goroutines and closes the records channel.
func RunFor(requests int, quit <-chan struct{}, concurrency int, url string, timeout time.Duration) <-chan request.Record {
	// sem is a semaphore to constrain access to at most n concurrent threads.
	sem := make(chan int, concurrency)
	rec := make(chan request.Record, requests)

	var wg sync.WaitGroup

	acquire := func() {
		sem <- 1
		wg.Add(1)
	}
	release := func() {
		<-sem
		wg.Done()
	}

	go func() {
		defer func() {
			wg.Wait()
			close(rec)
		}()
		for i := 0; i < requests; i++ {
			select {
			case <-quit:
				return
			default:
			}
			acquire()
			go func() {
				defer release()
				rec <- request.Do(url, timeout)
			}()
		}
	}()

	return rec
}
