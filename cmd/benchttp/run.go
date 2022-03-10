package main

import (
	"context"
	"errors"
	"flag"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/internal/configfile"
	"github.com/benchttp/runner/internal/configflags"
	"github.com/benchttp/runner/internal/signals"
	"github.com/benchttp/runner/output"
	"github.com/benchttp/runner/requester"
)

// cmdRun handles subcommand "benchttp run [options]".
type cmdRun struct {
	flagset *flag.FlagSet

	// configFile is the parsed value for flag -configFile
	configFile string

	// config is the runner config resulting from parsing CLI flags.
	config config.Global
}

// ensure cmdRun implements command
var _ command = (*cmdRun)(nil)

// init initializes cmdRun with default values.
func (cmd *cmdRun) init() {
	cmd.config = config.Default()
	cmd.configFile = configfile.Find([]string{
		"./.benchttp.yml",
		"./.benchttp.yaml",
		"./.benchttp.json",
	})
}

// execute runs the benchttp runner: it parses CLI flags, loads config
// from config file and parsed flags, then runs the benchmark and outputs
// it according to the config.
func (cmd *cmdRun) execute(args []string) error {
	cmd.init()

	fieldsSet := cmd.parseArgs(args)

	cfg, err := cmd.makeConfig(fieldsSet)
	if err != nil {
		return err
	}

	req, err := cfg.Request.Value()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go signals.ListenOSInterrupt(cancel)

	rep, err := requester.New(cmd.requesterConfig(cfg)).Run(ctx, req)
	if err != nil {
		if errors.Is(err, requester.ErrCanceled) {
			if err := cmd.handleRunInterrupt(); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return cmd.handleOutputError(output.New(rep, cfg).Export())
}

// parseArgs parses input args as config fields and returns
// a slice of fields that were set by the user.
func (cmd *cmdRun) parseArgs(args []string) []string {
	// first arg is subcommand "run"
	// skip parsing if no flags are provided
	if len(args) <= 1 {
		return []string{}
	}

	// config file path
	cmd.flagset.StringVar(&cmd.configFile,
		"configFile",
		cmd.configFile,
		"Config file path",
	)

	// attach config options flags to the flagset
	// and bind their value to the config struct
	configflags.Set(cmd.flagset, &cmd.config)

	cmd.flagset.Parse(args[1:]) //nolint:errcheck // never occurs due to flag.ExitOnError

	return configflags.Which(cmd.flagset)
}

// makeConfig returns a config.Global initialized with config file
// options if found, overridden with CLI options listed in fields
// slice param.
func (cmd *cmdRun) makeConfig(fields []string) (cfg config.Global, err error) {
	// configFile not set and default ones not found:
	// skip the merge and return the cli config
	if cmd.configFile == "" {
		return cmd.config, cmd.config.Validate()
	}

	fileConfig, err := configfile.Parse(cmd.configFile)
	if err != nil && !errors.Is(err, configfile.ErrFileNotFound) {
		// config file is not mandatory: discard ErrFileNotFound.
		// other errors are critical
		return
	}

	mergedConfig := fileConfig.Override(cmd.config, fields...)

	return mergedConfig, mergedConfig.Validate()
}

// requesterConfig returns a requester.Config generated from cfg.
func (*cmdRun) requesterConfig(cfg config.Global) requester.Config {
	return requester.Config{
		Requests:       cfg.Runner.Requests,
		Concurrency:    cfg.Runner.Concurrency,
		Interval:       cfg.Runner.Interval,
		RequestTimeout: cfg.Runner.RequestTimeout,
		GlobalTimeout:  cfg.Runner.GlobalTimeout,
		Silent:         cfg.Output.Silent,
	}
}

// handleRunInterrupt handles the case when the runner is interrupted.
func (*cmdRun) handleRunInterrupt() error {
	v, err := promptf("\nBenchmark interrupted, generate output anyway? (yes/no): ")
	if err != nil {
		return err
	}
	if v != "yes" {
		return errors.New("benchmark interrupted without output")
	}
	return nil
}

func (*cmdRun) handleOutputError(err error) error {
	if err == nil {
		return nil
	}
	if output.ExportErrorOf(err).HasAuthError() {
		return errors.New(
			"authentification to benchttp server failed, " +
				`please run "benchttp auth login" and restart the benchmark`,
		)
	}
	return err
}
