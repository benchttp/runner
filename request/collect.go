package request

import "github.com/benchttp/runner/record"

func collect(c <-chan record.Record) []record.Record {
	rec := []record.Record{}

	for r := range c {
		rec = append(rec, r)
	}
	return rec
}
