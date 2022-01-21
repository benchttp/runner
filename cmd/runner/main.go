package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/benchttp/runner/report"
	"github.com/benchttp/runner/request"
)

const (
	DefaultConcurrency = 1
	DefaultRequests    = 0 // Use duration as exit condition if omitted.
	DefaultDuration    = 60
	DefaultTimeout     = 10
)

var (
	url         string
	concurrency int           // Number of connections to run concurrently.
	requests    int           // Number of requests to run, use duration as exit condition if omitted.
	duration    time.Duration // Duration of test, in seconds.
	timeout     time.Duration // Timeout for each http request, in seconds.
)

func parseArgs() {
	c := flag.Int("c", DefaultConcurrency, "Number of connections to run concurrently")
	r := flag.Int("r", DefaultRequests, "Number of requests to run, use duration as exit condition if omitted")
	d := flag.Int("d", DefaultDuration, "Duration of test, in seconds")
	t := flag.Int("t", DefaultTimeout, "Timeout for each http request, in seconds")

	flag.Parse()

	url = os.Args[len(os.Args)-1]
	fmt.Printf("Testing url: %s\n", url)

	concurrency = *c
	fmt.Printf("concurrency: %d\n", concurrency)
	requests = *r
	if *r > 0 {
		fmt.Printf("requests: %d\n", requests)
	}
	duration = (time.Duration(*d)) * time.Second
	fmt.Printf("duration: %s\n", duration)
	timeout = (time.Duration(*t)) * time.Second

	println()
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
		rec = request.Do(requests, quit, concurrency, url, timeout)
	} else {
		rec = request.DoUntil(quit, concurrency, url, timeout)
	}

	reports := report.Collect(rec)
	println(len(reports))
}
