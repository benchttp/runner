package config_test

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
)

var (
	emptyBody, _ = config.NewBody("", "")
	validBody, _ = config.NewBody("application/json", "{\"key\": \"value\", \"key1\": \"value1\"}")
)

func TestValidate(t *testing.T) {
	t.Run("test valid configuration", func(t *testing.T) {
		cfg := config.Config{
			Request: config.Request{
				Timeout: 5,
				Body:    *validBody,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      5,
				Concurrency:   5,
				GlobalTimeout: 5,
			},
		}.WithURL("https://github.com/benchttp/")
		err := cfg.Validate()
		if err != nil {
			t.Errorf("valid configuration not considered as such")
		}
	})

	t.Run("test providing bodyType but no bodyContent value returns correct error", func(t *testing.T) {
		bodyWithoutContent, _ := config.NewBody("application/json", "")
		cfg := config.Config{
			Request: config.Request{
				Timeout: 5,
				Body:    *bodyWithoutContent,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      5,
				Concurrency:   5,
				GlobalTimeout: 5,
			},
		}.WithURL("github-com/benchttp")
		err := cfg.Validate()
		if err == nil {
			t.Errorf("bodyType provided without bodyContent considered valid")
		} else if !errorContains(err, "-bodyContent: you must provide a value if you have added a bodyType") {
			t.Errorf("\n- information about bodyType provided without bodyContent missing from error message")
		}
	})

	t.Run("test providing bodyContent but no bodyType value returns correct error", func(t *testing.T) {
		bodyWithoutType, _ := config.NewBody("", "{\"key\": \"value\", \"key1\": \"value1\"}")
		cfg := config.Config{
			Request: config.Request{
				Timeout: 5,
				Body:    *bodyWithoutType,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      5,
				Concurrency:   5,
				GlobalTimeout: 5,
			},
		}.WithURL("github-com/benchttp")
		err := cfg.Validate()
		if err == nil {
			t.Errorf("bodyContent provided without bodyType considered valid")
		} else if !errorContains(err, "-bodyType: you must provide a value if you have added a bodyContent") {
			t.Errorf("\n- information about bodyContent provided without bodyType missing from error message")
		}
	})

	t.Run("test invalid bodyType returns correct error", func(t *testing.T) {
		bodyWithInvalidType, _ := config.NewBody("invalid value", "{\"key\": \"value\", \"key1\": \"value1\"}")
		cfg := config.Config{
			Request: config.Request{
				Timeout: 5,
				Body:    *bodyWithInvalidType,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      5,
				Concurrency:   5,
				GlobalTimeout: 5,
			},
		}.WithURL("github-com/benchttp")
		err := cfg.Validate()
		if err == nil {
			t.Errorf("invalid bodyType not considered as such")
		} else if !errorContains(err, "bodyType: invalid value") {
			t.Errorf("\n- information about invalid bodyType missing from error message")
		}
	})

	t.Run("test invalid bodyContent returns correct error", func(t *testing.T) {
		_, err := config.NewBody("application/json", "{\"invalid json\"}")
		if err == nil {
			t.Errorf("invalid bodyContent not considered as such")
		} else if !errorContains(err, "bodyContent is not valid json data") {
			t.Errorf("\n- information about invalid bodyContent missing from error message")
		}
	})

	// Body is validated before Config.Validate() is used so we do not check it here and provide a valid empty body
	t.Run("test invalid configuration returns ErrInvalid error with correct messages", func(t *testing.T) {
		cfg := config.Config{
			Request: config.Request{
				Timeout: -5,
				Body:    *emptyBody,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      -5,
				Concurrency:   -5,
				GlobalTimeout: -5,
			},
		}.WithURL("github-com/benchttp")
		err := cfg.Validate()
		if err == nil {
			t.Errorf("invalid configuration considered valid")
		} else {
			if !errorContains(err, "-url: "+cfg.Request.URL.String()+" is not a valid url") {
				t.Errorf("\n- information about invalid url missing from error message")
			}
			if !errorContains(err, "-requests: must be >= 0, we got ") {
				t.Errorf("\n- information about invalid requests number missing from error message")
			}
			if !errorContains(err, "-concurrency: must be > 0, we got ") {
				t.Errorf("\n- information about invalid concurrency number missing from error message")
			}
			if !errorContains(err, "-timeout: must be > 0, we got") {
				t.Errorf("\n- information about invalid timeout missing from error message")
			}
			if !errorContains(err, "-globalTimeout: must be > 0, we got ") {
				t.Errorf("\n- information about invalid globalTimeout missing from error message")
			}
		}
	})
}

func TestWithURL(t *testing.T) {
	t.Run("set empty url if invalid", func(t *testing.T) {
		cfg := config.Config{}.WithURL("abc")
		if got := cfg.Request.URL; !reflect.DeepEqual(got, &url.URL{}) {
			t.Errorf("exp empty *url.URL, got %v", got)
		}
	})

	t.Run("set parsed url", func(t *testing.T) {
		var (
			rawURL    = "http://benchttp.app?cool=true"
			expURL, _ = url.ParseRequestURI(rawURL)
			gotURL    = config.Config{}.WithURL(rawURL).Request.URL
		)

		if !reflect.DeepEqual(gotURL, expURL) {
			t.Errorf("\nexp %v\ngot %v", expURL, gotURL)
		}
	})
}

func TestOverride(t *testing.T) {
	t.Run("do not override unspecified fields", func(t *testing.T) {
		baseCfg := config.Config{}
		newCfg := config.Config{
			Request: config.Request{
				Timeout: 3 * time.Second,
				Body:    *validBody,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      1,
				Concurrency:   2,
				GlobalTimeout: 4 * time.Second,
			},
		}.WithURL("http://a.b?p=2")

		if gotCfg := baseCfg.Override(newCfg); !reflect.DeepEqual(gotCfg, baseCfg) {
			t.Errorf("overrode unexpected fields:\nexp %#v\ngot %#v", baseCfg, gotCfg)
		}
	})

	t.Run("override specified fields", func(t *testing.T) {
		baseCfg := config.Config{}
		newCfg := config.Config{
			Request: config.Request{
				Timeout: 3 * time.Second,
				Body:    *validBody,
			},
			RunnerOptions: config.RunnerOptions{
				Requests:      1,
				Concurrency:   2,
				GlobalTimeout: 4 * time.Second,
			},
		}.WithURL("http://a.b?p=2")
		fields := []string{
			config.FieldMethod,
			config.FieldURL,
			config.FieldTimeout,
			config.FieldRequests,
			config.FieldConcurrency,
			config.FieldGlobalTimeout,
			config.FieldBodyType,
			config.FieldBodyContent,
		}

		if gotCfg := baseCfg.Override(newCfg, fields...); !reflect.DeepEqual(gotCfg, newCfg) {
			t.Errorf("did not override expected fields:\nexp %v\ngot %v", baseCfg, gotCfg)
			t.Log(fields)
		}
	})

	t.Run("override header selectively", func(t *testing.T) {
		testcases := []struct {
			label     string
			oldHeader http.Header
			newHeader http.Header
			expHeader http.Header
		}{
			{
				label:     "erase overridden keys",
				oldHeader: http.Header{"key": []string{"oldval"}},
				newHeader: http.Header{"key": []string{"newval"}},
				expHeader: http.Header{"key": []string{"newval"}},
			},
			{
				label:     "do not erase not overridden keys",
				oldHeader: http.Header{"key": []string{"oldval"}},
				newHeader: http.Header{},
				expHeader: http.Header{"key": []string{"oldval"}},
			},
			{
				label:     "add new keys",
				oldHeader: http.Header{"key0": []string{"oldval"}},
				newHeader: http.Header{"key1": []string{"newval"}},
				expHeader: http.Header{
					"key0": []string{"oldval"},
					"key1": []string{"newval"},
				},
			},
			{
				label: "erase only overridden keys",
				oldHeader: http.Header{
					"key0": []string{"oldval0", "oldval1"},
					"key1": []string{"oldval0", "oldval1"},
				},
				newHeader: http.Header{
					"key1": []string{"newval0", "newval1"},
					"key2": []string{"newval0", "newval1"},
				},
				expHeader: http.Header{
					"key0": []string{"oldval0", "oldval1"},
					"key1": []string{"newval0", "newval1"},
					"key2": []string{"newval0", "newval1"},
				},
			},
			{
				label:     "nil new header does nothing",
				oldHeader: http.Header{"key": []string{"val"}},
				newHeader: nil,
				expHeader: http.Header{"key": []string{"val"}},
			},
			{
				label:     "replace nil old header",
				oldHeader: nil,
				newHeader: http.Header{"key": []string{"val"}},
				expHeader: http.Header{"key": []string{"val"}},
			},
			{
				label:     "nil over nil is nil",
				oldHeader: nil,
				newHeader: nil,
				expHeader: nil,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				oldCfg := config.Config{
					Request: config.Request{
						Header: tc.oldHeader,
					},
				}

				newCfg := config.Config{
					Request: config.Request{
						Header: tc.newHeader,
					},
				}

				gotCfg := oldCfg.Override(newCfg, config.FieldHeader)

				if gotHeader := gotCfg.Request.Header; !reflect.DeepEqual(gotHeader, tc.expHeader) {
					t.Errorf("\nexp %#v\ngot %#v", tc.expHeader, gotHeader)
				}
			})
		}
	})
}

// To check that the error message is as expected
func errorContains(err error, expected string) bool {
	if err == nil {
		return expected == ""
	}
	if expected == "" {
		return false
	}
	return strings.Contains(err.Error(), expected)
}
