package config

import (
	"net/url"
	"time"
)

var defaultConfig = Config{
	Request: Request{
		Method: "GET",
		URL:    &url.URL{},
	},
	RunnerOptions: RunnerOptions{
		Concurrency:    1,
		Requests:       -1, // Use GlobalTimeout as exit condition.
		RequestTimeout: 10 * time.Second,
		GlobalTimeout:  30 * time.Second,
	},
}
