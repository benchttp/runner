package configparser

import (
	"net/url"
	"os"
	"path"
	"time"

	"github.com/benchttp/runner/config"
)

// Parse parses a benchttp runner config file into a config.Config
// and returns it or the first non-nil error occurring in the process.
func Parse(cfgpath string) (config.Config, error) {
	b, err := os.ReadFile(cfgpath)
	if err != nil {
		return config.Config{}, err
	}

	ext := extension(path.Ext(cfgpath))
	parser, err := newParser(ext)
	if err != nil {
		return config.Config{}, err
	}

	var cfg rawConfig
	if err := parser.ParseConfig(b, &cfg); err != nil {
		return config.Config{}, err
	}

	return parseRawConfig(cfg)
}

// parseRawConfig parses an input raw config as a config.Config and returns it
// or the first non-nil error occurring in the process.
func parseRawConfig(in rawConfig) (cfg config.Config, err error) {
	parsedURL, err := parseAndBuildURL(in.Request.URL, in.Request.QueryParams)
	if err != nil {
		return
	}

	parsedGlobalTimeout, err := parseOptionalDuration(in.RunnerOptions.GlobalTimeout)
	if err != nil {
		return
	}

	parsedRequestTimeout, err := parseOptionalDuration(in.RunnerOptions.RequestTimeout)
	if err != nil {
		return
	}

	cfg.Request.Method = in.Request.Method
	cfg.Request.URL = parsedURL
	cfg.RunnerOptions.Requests = in.RunnerOptions.Requests
	cfg.RunnerOptions.Concurrency = in.RunnerOptions.Concurrency
	cfg.RunnerOptions.GlobalTimeout = parsedGlobalTimeout
	cfg.RunnerOptions.RequestTimeout = parsedRequestTimeout

	return
}

// parseAndBuildURL parses a raw string as a *url.URL and adds any extra
// query parameters. It returns the first non-nil error occurring in the
// process.
func parseAndBuildURL(raw string, qp map[string]string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	// retrieve url query, add extra params, re-attach to url
	q := u.Query()
	for k, v := range qp {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	return u, nil
}

// parseOptionalDuration parses the raw string as a time.Duration
// and returns the parsed value or a non-nil error.
// Contrary to time.ParseDuration, it does not return an error
// if raw == "".
func parseOptionalDuration(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	return time.ParseDuration(raw)
}
