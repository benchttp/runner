package request

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func doRequest(url string, timeout time.Duration) Record {
	client := http.Client{
		// Timeout includes connection time, any redirects, and reading the response body.
		// We may want exclude reading the response body in our benchmark tool.
		Timeout: timeout,
	}

	t0 := time.Now()

	resp, err := client.Get(url)
	if err != nil {
		return Record{error: fmt.Sprint(err)}
	}

	return newRecord(resp, time.Since(t0))
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
// available and sends the results through a pipeline as they come in.
// Returns a channel to pipeline the records.
// The value of concurrency limits the number of concurrent threads.
// Once all requests have been made or on quit signal, waits for
// goroutines and closes the records channel.
func Do(requests int, quit <-chan struct{}, concurrency int, url string, timeout time.Duration) <-chan Record {
	// sem is a semaphore to constrain access to at most n concurrent threads.
	sem := make(chan int, concurrency)
	rec := make(chan Record, requests)

	var wg sync.WaitGroup

	for i := 0; i < requests; i++ {
		select {
		case <-quit:
			break
		default:
		}
		acquire(sem, &wg)
		go func() {
			defer release(sem, &wg)
			rec <- doRequest(url, timeout)
		}()
	}

	wg.Wait()
	close(rec)
	return rec
}

// DoUntil launches a goroutine to ping url as soon as a thread is
// available and sends the results through a pipeline as they come in.
// Returns a channel to pipeline the records.
// The value of concurrency limits the number of concurrent threads.
// On quit signal, waits for goroutines and closes the records channel.
func DoUntil(quit <-chan struct{}, concurrency int, url string, timeout time.Duration) <-chan Record {
	// sem is a semaphore to constrain access to at most n concurrent threads.
	sem := make(chan int, concurrency)
	rec := make(chan Record)

	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case <-quit:
				wg.Wait()
				close(rec)
			default:
			}
			acquire(sem, &wg)
			go func() {
				defer release(sem, &wg)
				rec <- doRequest(url, timeout)
			}()
		}
	}()

	return rec
}
