package report

import "github.com/benchttp/runner/request"

func Collect(c <-chan request.Record) []request.Record {
	rec := []request.Record{}

	for r := range c {
		rec = append(rec, r)
	}
	return rec
}
