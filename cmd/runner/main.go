package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/benchttp/runner/config"
	configfile "github.com/benchttp/runner/config/file"
	"github.com/benchttp/runner/requester"
)

const (
	// API server endpoint. May live is some config file later.
	reportURL = "http://localhost:9998/report"
)

var (
	configFile    string
	uri           string
	header        = http.Header{}
	concurrency   int           // Number of connections to run concurrently
	requests      int           // Number of requests to run, use duration as exit condition if omitted.
	timeout       time.Duration // Timeout for each HTTP request
	interval      time.Duration // Minimum duration between two groups of requests
	globalTimeout time.Duration // Duration of test
)

var defaultConfigFiles = []string{
	"./.benchttp.yml",
	"./.benchttp.yaml",
	"./.benchttp.json",
}

func parseArgs() {
	flag.StringVar(&configFile, "configFile", configfile.Find(defaultConfigFiles), "Config file path")
	flag.StringVar(&uri, "url", "", "Target URL to request")
	flag.Var(headerValue{header: &header}, "header", "HTTP request header")
	flag.IntVar(&concurrency, "concurrency", 0, "Number of connections to run concurrently")
	flag.IntVar(&requests, "requests", 0, "Number of requests to run, use duration as exit condition if omitted")
	flag.DurationVar(&timeout, "timeout", 0, "Timeout for each HTTP request")
	flag.DurationVar(&interval, "interval", 0, "Minimum duration between two non concurrent requests")
	flag.DurationVar(&globalTimeout, "globalTimeout", 0, "Duration of test")
	flag.Parse()
}

func main() {
	parseArgs()

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := requester.New(cfg).RunAndSendReport(reportURL); err != nil {
		log.Fatal(err)
	}
}

// parseConfig returns a config.Config initialized with config file
// options if found, overridden with CLI options.
func parseConfig() (config.Config, error) {
	fileCfg, err := configfile.Parse(configFile)
	if err != nil && !errors.Is(err, configfile.ErrFileNotFound) {
		// config file is not mandatory, other errors are critical
		log.Fatal(err)
	}

	cliCfg := config.Config{
		Request: config.Request{
			Header:  header,
			Timeout: timeout,
		},
		RunnerOptions: config.RunnerOptions{
			Requests:      requests,
			Concurrency:   concurrency,
			Interval:      interval,
			GlobalTimeout: globalTimeout,
		},
	}.WithURL(uri)

	mergedConfig := fileCfg.Override(cliCfg, configFlags()...)

	return mergedConfig, mergedConfig.Validate()
}

// configFlags returns a slice of all config fields set via the CLI.
func configFlags() []string {
	var fields []string
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if name := f.Name; config.IsField(name) {
			fields = append(fields, name)
		}
	})
	return fields
}

type headerValue struct {
	header *http.Header
}

func (v headerValue) String() string {
	return fmt.Sprint(v.header)
}

func (v headerValue) Set(in string) error {
	fmt.Println("hello ", in)
	keyval := strings.Split(in, ":")
	if len(keyval) != 2 {
		return errors.New(`expect format key:value`)
	}
	key, val := keyval[0], keyval[1]
	(*v.header)[key] = append((*v.header)[key], val)
	return nil
}
