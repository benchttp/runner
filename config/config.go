package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Body struct {
	Type    string
	Content string
}

func NewBody(bodyType, bodyContent string) Body {
	var body Body
	body.Type = bodyType
	body.Content = bodyContent
	return body
}

var contentTypeValidValues []string = []string{"application/json"}

// Request contains the confing options relative to a single request.
type Request struct {
	Method  string
	URL     *url.URL
	Header  http.Header
	Timeout time.Duration
	Body    Body
}

// RunnerOptions contains options relative to the runner.
type RunnerOptions struct {
	Requests      int
	Concurrency   int
	Interval      time.Duration
	GlobalTimeout time.Duration
}

// Config represents the configuration of the runner.
// It must be validated using Config.Validate before usage.
type Config struct {
	Request       Request
	RunnerOptions RunnerOptions
}

// String returns an indented JSON representation of Config
// for debugging purposes.
func (cfg Config) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}

// HTTPRequest returns a *http.Request created from Target. Returns any non-nil
// error that occurred.
func (cfg Config) HTTPRequest() (*http.Request, error) {
	if cfg.Request.URL == nil {
		return nil, errors.New("empty url")
	}
	rawURL := cfg.Request.URL.String()
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return nil, errors.New("bad url")
	}
	// TODO: handle body
	req, err := http.NewRequest(cfg.Request.Method, rawURL, strings.NewReader(cfg.Request.Body.Content))
	if err != nil {
		return nil, err
	}
	req.Header = cfg.Request.Header
	return req, nil
}

// Override returns a new Config based on cfg with overridden values from c.
// Only fields specified in options are replaced. Accepted options are limited
// to existing Fields, other values are silently ignored.
func (cfg Config) Override(c Config, fields ...string) Config {
	for _, field := range fields {
		switch field {
		case FieldMethod:
			cfg.Request.Method = c.Request.Method
		case FieldURL:
			cfg.Request.URL = c.Request.URL
		case FieldHeader:
			cfg.overrideHeader(c.Request.Header)
		case FieldTimeout:
			cfg.Request.Timeout = c.Request.Timeout
		case FieldRequests:
			cfg.RunnerOptions.Requests = c.RunnerOptions.Requests
		case FieldConcurrency:
			cfg.RunnerOptions.Concurrency = c.RunnerOptions.Concurrency
		case FieldInterval:
			cfg.RunnerOptions.Interval = c.RunnerOptions.Interval
		case FieldGlobalTimeout:
			cfg.RunnerOptions.GlobalTimeout = c.RunnerOptions.GlobalTimeout
		}
	}
	return cfg
}

func (cfg *Config) overrideHeader(newHeader http.Header) {
	if newHeader == nil {
		return
	}
	if cfg.Request.Header == nil {
		cfg.Request.Header = http.Header{}
	}
	for k, v := range newHeader {
		cfg.Request.Header[k] = v
	}
}

// WithURL sets the current Config to the parsed *url.URL from rawURL
// and returns it. Any errors is discarded as a Config can be invalid
// until Config.Validate is called. The url is guaranteed not to be nil.
func (cfg Config) WithURL(rawURL string) Config {
	// ignore err: a Config can be invalid at this point
	urlURL, _ := url.ParseRequestURI(rawURL)
	if urlURL == nil {
		urlURL = &url.URL{}
	}
	cfg.Request.URL = urlURL
	return cfg
}

// Validate returns the config and a not nil ErrInvalid if any of the fields provided by the user is not valid
func (cfg Config) Validate() error { //nolint:gocognit
	inputErrors := []error{}

	if cfg.Request.URL == nil {
		inputErrors = append(inputErrors, errors.New("-url: missing url"))
	} else if _, err := url.ParseRequestURI(cfg.Request.URL.String()); err != nil {
		inputErrors = append(inputErrors, fmt.Errorf("-url: %s is not a valid url", cfg.Request.URL.String()))
	}

	if cfg.RunnerOptions.Requests < 1 && cfg.RunnerOptions.Requests != -1 {
		inputErrors = append(inputErrors, fmt.Errorf("-requests: must be >= 0, we got %d", cfg.RunnerOptions.Requests))
	}

	if cfg.RunnerOptions.Concurrency < 1 && cfg.RunnerOptions.Concurrency != -1 {
		inputErrors = append(inputErrors, fmt.Errorf("-concurrency: must be > 0, we got %d", cfg.RunnerOptions.Concurrency))
	}

	if cfg.Request.Timeout < 0 {
		inputErrors = append(inputErrors, fmt.Errorf("-timeout: must be > 0, we got %d", cfg.Request.Timeout))
	}

	if cfg.RunnerOptions.Interval < 0 {
		inputErrors = append(inputErrors, fmt.Errorf("-interval: must be > 0, we got %d", cfg.RunnerOptions.Interval))
	}

	if cfg.RunnerOptions.GlobalTimeout < 0 {
		inputErrors = append(inputErrors, fmt.Errorf("-globalTimeout: must be > 0, we got %d", cfg.RunnerOptions.GlobalTimeout))
	}

	if !reflect.ValueOf(cfg.Request.Body).IsZero() {
		if reflect.ValueOf(cfg.Request.Body.Type).IsZero() {
			inputErrors = append(inputErrors, fmt.Errorf("-bodyType: you must provide a value if you have added a bodyContent"))
		}
		if !contains(contentTypeValidValues, cfg.Request.Body.Type) {
			inputErrors = append(inputErrors, fmt.Errorf("-bodyType: invalid value, we got %s", cfg.Request.Body.Type))
		}
		if reflect.ValueOf(cfg.Request.Body.Content).IsZero() {
			inputErrors = append(inputErrors, fmt.Errorf("-bodyContent: you must provide a value if you have added a bodyType"))
		}
		if !isJSON(cfg.Request.Body.Content) {
			inputErrors = append(inputErrors, fmt.Errorf("-bodyContent: it is not valid json, we got %s", cfg.Request.Body.Content))
		}
	}

	if len(inputErrors) > 0 {
		return &ErrInvalid{inputErrors}
	}

	return nil
}

// Default returns a default config that is safe to use.
func Default() Config {
	return defaultConfig
}

// helpers
// Check that a value is present in a string slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Check that a string is valid JSON
func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}
