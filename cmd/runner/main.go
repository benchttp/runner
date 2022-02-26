package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/internal/configfile"
	"github.com/benchttp/runner/internal/configflags"
	"github.com/benchttp/runner/output"
	"github.com/benchttp/runner/requester"
)

var (
	configFile string

	cliConfig config.Global
)

var defaultConfigFiles = []string{
	"./.benchttp.yml",
	"./.benchttp.yaml",
	"./.benchttp.json",
}

func parseRunFlags() []string {
	flagset := flag.NewFlagSet("run", flag.ExitOnError)

	// config file path
	flagset.StringVar(&configFile,
		"configFile",
		configfile.Find(defaultConfigFiles),
		"Config file path",
	)

	// cli config
	configflags.Set(flagset, &cliConfig)

	if err := flagset.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	return configflags.Which(flagset)
}

func main() {
	// benchttp run [options]
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// TODO: benchttp auth [options]
}

func run() error {
	fieldsSet := parseRunFlags()

	cfg, err := makeConfig(fieldsSet)
	if err != nil {
		return err
	}

	req, err := cfg.Request.Value()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go listenOSInterrupt(cancel)

	rep, err := requester.New(requesterConfig(cfg)).Run(ctx, req)
	if err != nil {
		if errors.Is(err, requester.ErrCanceled) {
			if err := handleRunInterrupt(); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return output.New(rep, cfg).Export()
}

// makeConfig returns a config.Config initialized with config file
// options if found, overridden with CLI options listed in fields
// slice param.
func makeConfig(fields []string) (cfg config.Global, err error) {
	fileConfig, err := configfile.Parse(configFile)
	if err != nil && !errors.Is(err, configfile.ErrFileNotFound) {
		// config file is not mandatory, other errors are critical
		return
	}

	mergedConfig := fileConfig.Override(cliConfig, fields...)

	return mergedConfig, mergedConfig.Validate()
}

// requesterConfig returns a requester.Config generated from cfg.
func requesterConfig(cfg config.Global) requester.Config {
	return requester.Config{
		Requests:       cfg.Runner.Requests,
		Concurrency:    cfg.Runner.Concurrency,
		Interval:       cfg.Runner.Interval,
		RequestTimeout: cfg.Runner.RequestTimeout,
		GlobalTimeout:  cfg.Runner.GlobalTimeout,
		Silent:         cfg.Output.Silent,
	}
}

// listenOSInterrupt listens for OS interrupt signals and calls callback.
// It should be called in a separate goroutine from main as it blocks
// the execution until the OS interrupt signal is received.
func listenOSInterrupt(callback func()) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	<-sigC
	callback()
}

// handleRunInterrupt handles the case when the runner is interrupted.
func handleRunInterrupt() error {
	reader := bufio.NewReader(os.Stdin)
	// TODO: list output strategies
	// TODO: do not prompt if strategy is stdout only
	// TODO: add config option "output.generateOnCancel" and remove prompt?
	fmt.Printf("\nBenchmark interrupted, generate output anyway? (yes/no): ")
	line, _, err := reader.ReadLine()
	if err != nil {
		return err
	}
	if string(line) != "yes" {
		return errors.New("benchmark interrupted without output")
	}
	return nil
}
