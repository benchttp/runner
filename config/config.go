package config

import (
	"encoding/json"
	"errors"
	"net/url"
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

// Merge returns a zero Config.
//
// Once implemented, Merge will apply an override Config over a base Config.
func Merge(base, override Config) Config {
	return Config{}
}

// Validate returns an unimplemented error.
//
// Once implemented, Validate will return ErrInvalid if any of its fields
// does not meet the runner requirements.
func Validate() error {
	return errors.New("unimplemented")
}
