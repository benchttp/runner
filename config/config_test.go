package config_test

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
)

func TestNew(t *testing.T) {
	t.Run("zero value params return empty config", func(t *testing.T) {
		exp := config.Config{}
		if got := config.New("", 0, 0, 0, 0); !reflect.DeepEqual(got, exp) {
			t.Errorf("returned non-zero config:\nexp %#v\ngot %#v", exp, got)
		}
	})

	t.Run("non-zero params return initialized config", func(t *testing.T) {
		var (
			rawURL      = "http://example.com"
			urlURL, _   = url.Parse(rawURL)
			requests    = 1
			concurrency = 2
			reqTimeout  = 3 * time.Second
			glbTimeout  = 4 * time.Second
		)

		exp := config.Config{
			Request: config.Request{
				Method:  "",
				URL:     urlURL,
				Timeout: reqTimeout,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      requests,
				Concurrency:   concurrency,
				GlobalTimeout: glbTimeout,
			},
		}

		got := config.New(rawURL, requests, concurrency, reqTimeout, glbTimeout)

		if !reflect.DeepEqual(got, exp) {
			t.Errorf("returned unexpected config:\nexp %#v\ngot %#v", exp, got)
		}
	})
}
