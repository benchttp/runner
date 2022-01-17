package sem

import (
	"sync"
	"time"

	"github.com/benchttp/runner/request"
)

// RunUntil launches a goroutine to ping url as soon as a thread is available
// and processes the results as they come in. The number of concurrent threads
// is limited by concurrency value. On quit signal, returns the results.
func RunUntil(quit <-chan struct{}, concurrency int, url string, timeout time.Duration) []request.Record {
	// sem is a semaphore to constrain access to at most n concurrent threads.
	sem := make(chan int, concurrency)
	c := make(chan request.Record)

	var wg sync.WaitGroup

	acquire := func() {
		sem <- 1
		wg.Add(1)
	}
	release := func() {
		<-sem
		wg.Done()
	}

	rec := []request.Record{}
	// done signals when the processing of rec is done.
	done := make(chan struct{}, 1)

	go func() {
		defer func() {
			done <- struct{}{}
		}()
		for r := range c {
			rec = append(rec, r)
		}
	}()

	for {
		select {
		case <-quit:
			wg.Wait()
			close(c)
			<-done // Block until c has been emptied.
			return rec
		default:
		}
		acquire()
		go func() {
			defer release()
			c <- request.Do(url, timeout)
		}()
	}
}
