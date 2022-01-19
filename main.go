package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/benchttp/runner/report"
	"github.com/benchttp/runner/request"
	"github.com/benchttp/runner/sem"
)

const (
	DEFAULT_CONCURRENCY = 1
	DEFAULT_REQUESTS    = 0 // Use duration as exit condition if omitted.
	DEFAULT_DURATION    = 60
	DEFAULT_TIMEOUT     = 10
)

var (
	url         string
	concurrency int           // Number of connections to run concurrently.
	requests    int           // Number of requests to run, use duration as exit condition if omitted.
	duration    time.Duration // Duration of test, in seconds.
	timeout     time.Duration // Timeout for each http request, in seconds.
)

func parseArgs() {
	c := flag.Int("c", DEFAULT_CONCURRENCY, "Number of connections to run concurrently")
	r := flag.Int("r", DEFAULT_REQUESTS, "Number of requests to run, use duration as exit condition if omitted")
	d := flag.Int("d", DEFAULT_DURATION, "Duration of test, in seconds")
	t := flag.Int("t", DEFAULT_TIMEOUT, "Timeout for each http request, in seconds")

	flag.Parse()

	concurrency = *c
	fmt.Printf("concurrency: %d\n", concurrency)
	requests = *r
	fmt.Printf("requests: %d\n", requests)
	duration = (time.Duration(*d)) * time.Second
	fmt.Printf("duration: %s\n", duration)
	timeout = (time.Duration(*t)) * time.Second

	url = os.Args[len(os.Args)-1]
	fmt.Printf("url: %s\n", url)
}

func main() {
	parseArgs()

	quit := make(chan struct{}, 1)
	defer close(quit)

	go func() {
		<-time.NewTimer(duration).C
		quit <- struct{}{}
	}()

	var rec <-chan request.Record
	if requests > 0 {
		rec = sem.RunFor(requests, quit, concurrency, url, timeout)
	} else {
		rec = sem.RunUntil(quit, concurrency, url, timeout)
	}

	reports := report.Collect(rec)
	println(len(reports))
}
