package main

import (
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

var (
	configFile  string
	url         string
	concurrency int           // Number of connections to run concurrently
	requests    int           // Number of requests to run, use duration as exit condition if omitted.
	duration    time.Duration // Duration of test
	timeout     time.Duration // Timeout for each http request
)

var defaultConfigFiles = []string{
	"./.benchttp.yml",
	"./.benchttp.yaml",
	"./.benchttp.json",
}

func parseArgs() {
	flag.StringVar(&configFile, "config-file", findFile(defaultConfigFiles), "Config file path")
	flag.StringVar(&url, "url", "", "Target URL to request")
	flag.IntVar(&concurrency, "c", 0, "Number of connections to run concurrently")
	flag.IntVar(&requests, "r", 0, "Number of requests to run, use duration as exit condition if omitted")
	flag.DurationVar(&duration, "d", 0, "Duration of test")
	flag.DurationVar(&timeout, "t", 0, "Timeout for each http request")
	flag.Parse()
}

func main() {
	parseArgs()

	cfg := makeRunnerConfig()
	fmt.Println(cfg)

	rec := request.Do(cfg)

	fmt.Println("total:", len(rec))
}

// makeRunnerConfig retrieves a config from
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

// findFile returns the first name tham matches a file path.
// If no match is found, it returns an empty string.
func findFile(names []string) string {
	for _, path := range names {
		if _, err := os.Stat(path); err == nil { // err IS nil
			return path
		}
	}
	return ""
}
