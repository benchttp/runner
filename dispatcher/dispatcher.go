package dispatcher

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

type Dispatcher interface {
	Do(ctx context.Context, maxIter int, callback func())
}

type dispatcher struct {
	sem *semaphore.Weighted
}

// Do concurrently executes callback at most maxIter times or until ctx is done
// or canceled. Concurrency is handled leveraging the semaphore pattern, which
// ensures at most Dispatcher.numWorkers goroutines are spawned at the same time.
func (d dispatcher) Do(ctx context.Context, maxIter int, callback func()) {
	maxIter = sanitizeMaxIter(maxIter)
	callback = sanitizeCallback(callback)

	wg := sync.WaitGroup{}

	for i := 0; i < maxIter || maxIter == 0; i++ {
		wg.Add(1)

		if err := d.sem.Acquire(ctx, 1); err != nil {
			// err is either context.DeadlineExceeded or context.Canceled
			// which are expected values so we stop the process silently.
			wg.Done()
			break
		}

		go func() {
			defer func() {
				d.sem.Release(1)
				wg.Done()
			}()
			callback()
		}()
	}

	wg.Wait()
}

// New returns a Dispatcher initialized with numWorker.
func New(numWorker int) Dispatcher {
	numWorker = sanitizeNumWorker(numWorker)
	sem := semaphore.NewWeighted(int64(numWorker))
	return dispatcher{sem: sem}
}

func sanitizeNumWorker(numWorkers int) int {
	if numWorkers < 1 {
		return 1
	}
	return numWorkers
}

func sanitizeMaxIter(maxIter int) int {
	if maxIter < 0 {
		return 0
	}
	return maxIter
}

func sanitizeCallback(callback func()) func() {
	if callback == nil {
		return func() {}
	}
	return callback
}
