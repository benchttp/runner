package config

import (
	"net/http"
	"net/url"
	"time"
)

var defaultConfig = Global{
	Request: Request{
		Method: "GET",
		URL:    &url.URL{},
		Header: http.Header{},
		Body:   Body{},
	},
	Runner: Runner{
		Concurrency:    1,
		Requests:       -1, // Use GlobalTimeout as exit condition.
		Interval:       0 * time.Second,
		RequestTimeout: 10 * time.Second,
		GlobalTimeout:  30 * time.Second,
	},
}
