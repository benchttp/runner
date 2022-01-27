package configparser_test

import (
	"net/url"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/configparser"
)

const (
	testdataConfigPath = "../test/testdata/config"
	testURL            = "http://localhost:9999?fib=30&delay=200ms"
)

var supportedExt = []string{
	".yml",
	".yaml",
	".json",
}

// TestParse ensures the config file is open, read, and correctly parsed.
func TestParse(t *testing.T) {
	t.Run("happy path for all extensions", func(t *testing.T) {
		for _, ext := range supportedExt {
			expCfg := newExpConfig()
			fname := path.Join(testdataConfigPath, "benchttp"+ext)

			gotCfg, err := configparser.Parse(fname)
			if err != nil {
				// critical error, stop the test
				t.Fatal(err)
			}

			gotURL, expURL := gotCfg.Request.URL, expCfg.Request.URL

			// compare *url.URLs separately, as they contain unpredictable values
			// they need special checks
			if !sameURL(gotCfg.Request.URL, expCfg.Request.URL) {
				t.Errorf(
					"unexpected parsed URL: exp %v, got %v",
					expCfg.Request.URL, gotCfg.Request.URL,
				)
			}

			// replace unpredictable values (undetermined query params order)
			restoreGotCfg := setTempValue(&gotURL.RawQuery, "<replaced by test>")
			restoreExpCfg := setTempValue(&expURL.RawQuery, "<replaced by test>")

			if !reflect.DeepEqual(gotCfg, expCfg) {
				t.Errorf("unexpected parsed config: exp %s\ngot %s", expCfg, gotCfg)
			}

			restoreExpCfg()
			restoreGotCfg()
		}
	})
}

// helpers

// newExpConfig returns the expected config.Config result after parsing
// one of the config files in testdataConfigPath.
func newExpConfig() config.Config {
	u, _ := url.Parse(testURL)
	return config.Config{
		Request: config.Request{
			Method: "GET",
			URL:    u,
		},

		RunnerOptions: config.RunnerOptions{
			Requests:       100,
			Concurrency:    1,
			GlobalTimeout:  60 * time.Second,
			RequestTimeout: 2 * time.Second,
		},
	}
}

// sameURL returns true if a and b are the same *url.URL, taking into account
// the undeterministic nature of their RawQuery.
func sameURL(a, b *url.URL) bool {
	// check query params equality via Query() rather than RawQuery
	if !reflect.DeepEqual(a.Query(), b.Query()) {
		return false
	}

	// temporarily set RawQuery to a determined value
	for _, u := range []*url.URL{a, b} {
		defer setTempValue(&u.RawQuery, "<replaced by test>")()
	}

	// we can now rely on deep equality check
	return reflect.DeepEqual(a, b)
}

// setTempValue sets *ptr to val and returns a restore func that sets *ptr
// back to its previous value.
func setTempValue(ptr *string, val string) (restore func()) {
	previousValue := *ptr
	*ptr = val
	return func() {
		*ptr = previousValue
	}
}
