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

const testdataConfigPath = "../test/testdata/config"

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
				// critical error, do not continue the test suite
				t.Fatal(err)
			}

			// compare *url.URLs separately, as they contain unpredictable values
			// they need special treatment
			if !sameURL(gotCfg.Request.URL, expCfg.Request.URL) {
				t.Errorf(
					"unexpected parsed URL: exp %v, got %v",
					expCfg.Request.URL, gotCfg.Request.URL,
				)
			}

			// replace unpredictable values (undetermined query params order)
			gotCfg.Request.URL.RawQuery = "replaced by test"
			expCfg.Request.URL.RawQuery = "replaced by test"

			if !reflect.DeepEqual(gotCfg, expCfg) {
				t.Errorf("unexpected parsed config: exp %s\ngot %s", expCfg, gotCfg)
			}
		}
	})
}

// newExpConfig returns the expected config.Config result after parsing
// one of the config files in testdataConfigPath.
func newExpConfig() config.Config {
	u, _ := url.Parse("http://localhost:9999?fib=30&delay=200ms")
	return config.Config{
		Request: struct {
			Method string
			URL    *url.URL
		}{
			Method: "GET",
			URL:    u,
		},

		RunnerOptions: struct {
			Requests       int
			Concurrency    int
			GlobalTimeout  time.Duration
			RequestTimeout time.Duration
		}{
			Requests:       100,
			Concurrency:    1,
			GlobalTimeout:  60 * time.Second,
			RequestTimeout: 2 * time.Second,
		},
	}
}

func sameURL(a, b *url.URL) bool {
	if !reflect.DeepEqual(a.Query(), b.Query()) {
		return false
	}

	rqa, rqb := a.RawQuery, b.RawQuery
	defer func() {
		a.RawQuery = rqa
		b.RawQuery = rqb
	}()
	a.RawQuery = ""
	b.RawQuery = ""

	return reflect.DeepEqual(a, b)
}
