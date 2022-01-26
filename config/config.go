package config

import (
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"time"
)

// ErrInvalid error is returned in the case of an invalid config.
var ErrInvalid = errors.New("invalid config")

// Config represents the configuration of the runner.
type Config struct {
	Request struct {
		Method string
		URL    *url.URL
	}

	RunnerOptions struct {
		Requests       int
		Concurrency    int
		GlobalTimeout  time.Duration
		RequestTimeout time.Duration
	}
}

// String returns an indented JSON representations of Config
// for debugging purposes.
func (cfg Config) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}

// New returns a Config initialized with given parameters.
func New(uri string, requests, concurrency int, requestTimeout, globalTimeout time.Duration) Config {
	var cfg Config
	cfg.Request.URL, _ = url.Parse(uri) // TODO: error handling
	cfg.RunnerOptions.Requests = requests
	cfg.RunnerOptions.Concurrency = concurrency
	cfg.RunnerOptions.GlobalTimeout = globalTimeout
	cfg.RunnerOptions.RequestTimeout = requestTimeout
	return cfg
}

// Merge returns a zero Config.
//
// Once implemented, Merge will apply an override Config over a base Config.
func Merge(base, override Config) Config {
	if override.Request.Method != "" {
		base.Request.Method = override.Request.Method
	}
	if reflect.ValueOf(override.Request.URL).IsZero() {
		base.Request.URL = override.Request.URL
	}
	if override.RunnerOptions.Requests != 0 {
		base.RunnerOptions.Requests = override.RunnerOptions.Requests
	}
	if override.RunnerOptions.Concurrency != 0 {
		base.RunnerOptions.Concurrency = override.RunnerOptions.Concurrency
	}
	if override.RunnerOptions.GlobalTimeout != 0 {
		base.RunnerOptions.GlobalTimeout = override.RunnerOptions.GlobalTimeout
	}
	if override.RunnerOptions.RequestTimeout != 0 {
		base.RunnerOptions.RequestTimeout = override.RunnerOptions.RequestTimeout
	}
	return base
}

// Validate returns an unimplemented error.
//
// Once implemented, Validate will return ErrInvalid if any of its fields
// does not meet the runner requirements.
func Validate() error {
	return errors.New("unimplemented")
}
