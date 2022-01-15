package sem

import (
	"sync"
	"time"

	"github.com/benchttp/runner/request"
)

// Run launches a goroutine to ping url as soon as a thread is available.
// The number of threads is limited by concurrency value.
// Once all requests have been made or on quit signal, the results are
// processed and returned.
func Run(quit <-chan struct{}, url string, concurrency int, requests int, timeout time.Duration) []request.Record {
	// sem is a semaphore to constrain access to at most n concurrent threads.
	sem := make(chan int, concurrency)
	c := make(chan request.Record, requests)

	var wg sync.WaitGroup

	acquire := func() {
		sem <- 1
		wg.Add(1)
	}
	release := func() {
		<-sem
		wg.Done()
	}

	for i := 0; i < requests; i++ {
		select {
		case <-quit:
			wg.Wait()
			close(c)
			return consume(c)
		default:
		}
		acquire()
		go func() {
			defer release()
			c <- request.Do(url, timeout)
		}()
	}

	wg.Wait()
	close(c)
	return consume(c)
}

func consume(c <-chan request.Record) []request.Record {
	rec := []request.Record{}
	for r := range c {
		rec = append(rec, r)
	}
	return rec
}
