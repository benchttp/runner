package semimpl_test

import (
	"context"
	"testing"
	"time"

	"github.com/benchttp/runner/semimpl"
)

func TestDo(t *testing.T) {
	t.Run("stop until maxIter is reached", func(t *testing.T) {
		var (
			numWorkers = 1
			gotIter    = 0
			maxIter    = 10
			expIter    = 10
		)

		semimpl.Do(context.Background(), numWorkers, maxIter, func() {
			gotIter++
		})

		if gotIter != expIter {
			t.Errorf("iterations: exp %d, got %d", expIter, gotIter)
		}
	})

	t.Run("stop on context timeout", func(t *testing.T) {
		var (
			timeout    = 50 * time.Millisecond
			interval   = timeout / 5
			numWorkers = 1
			maxIter    = 0 // infinite iterations

			margin      = 15 * time.Millisecond
			maxDuration = timeout + margin
		)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		gotDuration := timeFunc(func() {
			semimpl.Do(ctx, numWorkers, maxIter, func() {
				time.Sleep(interval)
			})
		})

		if gotDuration > maxDuration {
			t.Errorf(
				"context timeout duration: exp < %dms, got %dms",
				maxDuration.Milliseconds(), gotDuration.Milliseconds(),
			)
		}
	})

	t.Run("stop on context cancel", func(t *testing.T) {
		var (
			timeout    = 50 * time.Millisecond
			interval   = 10 * time.Millisecond
			numWorkers = 1
			maxIter    = 0 // infinite iterations

			margin      = 15 * time.Millisecond
			maxDuration = timeout + margin
		)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(timeout)
			cancel()
		}()

		gotDuration := timeFunc(func() {
			semimpl.Do(ctx, numWorkers, maxIter, func() {
				time.Sleep(interval)
			})
		})

		if gotDuration > maxDuration {
			t.Errorf(
				"context cancel duration: exp < %dms, got %dms",
				maxDuration.Milliseconds(), gotDuration.Milliseconds(),
			)
		}
	})

	t.Run("limit concurrent workers", func(t *testing.T) {
		var (
			numWorkers = 3
			maxIter    = 12

			durations                = make([]time.Duration, 0, maxIter)
			maxIntervalWithinGroup   = 10 * time.Millisecond
			minIntervalBetweenGroups = 30 * time.Millisecond
		)

		start := time.Now()
		semimpl.Do(context.Background(), numWorkers, maxIter, func() {
			durations = append(durations, time.Since(start))
			time.Sleep(minIntervalBetweenGroups)
		})

		// check duration slice is coherent, grouping its values
		// by expectedly similar durations, e.g.:
		// 12 iterations / 3 workers -> 4 groups of 3 similar durations.
		// With a callback duration of 30ms, we can expect something like:
		// [[0ms, 0ms, 0ms], [30ms, 30ms, 30ms], [60ms, 60ms, 60ms], [90ms, 90ms, 90ms]]
		// with a certain delta.
		groups := groupby(durations, numWorkers)
		for groupIndex, group := range groups {
			// check durations within each group are similar
			hi, lo := maxof(group), minof(group)
			if interval := hi - lo; interval > maxIntervalWithinGroup {
				t.Errorf(
					"unexpected interval in group: exp < %dms, got %dms",
					maxIntervalWithinGroup.Milliseconds(), interval.Milliseconds(),
				)
				t.Log(durations)
			}

			// check durations between distinct groups are spaced
			if groupIndex == len(groups)-1 {
				break
			}
			curr, next := minof(group), maxof(groups[groupIndex+1])
			if interval := next - curr; interval < minIntervalBetweenGroups {
				t.Errorf(
					"unexpected interval between groups: exp > %dms, got %dms",
					minIntervalBetweenGroups.Milliseconds(), interval.Milliseconds(),
				)
				t.Log(durations)
			}
		}
	})
}

// helpers

func groupby(src []time.Duration, by int) [][]time.Duration {
	numGroups := len(src) / by
	out := make([][]time.Duration, 0, numGroups)

	for i := 0; i < numGroups; i++ {
		lo := by * i
		hi := lo + by
		out = append(out, src[lo:hi])
	}

	return out
}

func minof(src []time.Duration) time.Duration {
	var min time.Duration
	for _, d := range src {
		if d < min || min == 0 {
			min = d
		}
	}
	return min
}

func maxof(src []time.Duration) time.Duration {
	var max time.Duration
	for _, d := range src {
		if d > max {
			max = d
		}
	}
	return max
}

func timeFunc(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}
