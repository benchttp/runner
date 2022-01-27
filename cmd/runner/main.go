package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/configparser"
	"github.com/benchttp/runner/request"
)

const (
	DefaultConfigFile  = "./.benchttp.yml"
	DefaultURL         = ""
	DefaultConcurrency = 1
	DefaultRequests    = 0 // Use duration as exit condition if omitted.
	DefaultDuration    = 60 * time.Second
	DefaultTimeout     = 10 * time.Second
)

// TODO: rethink defaulting process
var (
	configFile  string
	url         string
	concurrency int           // Number of connections to run concurrently
	requests    int           // Number of requests to run, use duration as exit condition if omitted.
	duration    time.Duration // Duration of test
	timeout     time.Duration // Timeout for each http request
)

func parseArgs() {
	flag.StringVar(&configFile, "config-file", DefaultConfigFile, "Config file path")
	flag.StringVar(&url, "url", DefaultURL, "Target URL to request")
	flag.IntVar(&concurrency, "c", DefaultConcurrency, "Number of connections to run concurrently")
	flag.IntVar(&requests, "r", DefaultRequests, "Number of requests to run, use duration as exit condition if omitted")
	flag.DurationVar(&duration, "d", DefaultDuration, "Duration of test")
	flag.DurationVar(&timeout, "t", DefaultTimeout, "Timeout for each http request")
	flag.Parse()
}

func main() {
	parseArgs()

	cfg := makeRunnerConfig()
	fmt.Println(cfg)

	// TODO: delay timeout creation
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// TODO: accept a config.Config struct
	rec := request.Do(ctx,
		cfg.RunnerOptions.Requests,
		cfg.RunnerOptions.Concurrency,
		cfg.Request.URL.String(),
		cfg.Request.Timeout,
	)

	fmt.Println("total:", len(rec))
}

func makeRunnerConfig() config.Config {
	fileConfig, err := configparser.Parse(configFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// config file is not mandatory, other errors are critical
		log.Fatal(err)
	}

	cliConfig := config.New(url, requests, concurrency, timeout, duration)

	cfg := config.Merge(fileConfig, cliConfig)

	return cfg
}
