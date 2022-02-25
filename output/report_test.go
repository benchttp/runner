package output_test

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/output"
	"github.com/benchttp/runner/requester"
)

func TestReport_String(t *testing.T) {
	t.Run("return default summary if template is empty", func(t *testing.T) {
		const tpl = ""

		rep := output.New(newBenchmark(), newConfig(tpl))
		checkSummary(t, rep.String())
	})

	t.Run("return executed template if valid", func(t *testing.T) {
		const tpl = "{{ .Benchmark.Length }}"

		bk := newBenchmark()
		rep := output.New(bk, newConfig(tpl))

		if got, exp := rep.String(), strconv.Itoa(bk.Length); got != exp {
			t.Errorf("\nunexpected output\nexp %s\ngot %s", exp, got)
		}
	})

	t.Run("fallback to default summary if template is invalid", func(t *testing.T) {
		const tpl = "{{ .Marcel.Patulacci }}"

		rep := output.New(newBenchmark(), newConfig(tpl))
		got := rep.String()
		split := strings.Split(got, "Falling back to default summary:\n")

		if len(split) != 2 {
			t.Fatalf("\nunexpected output:\n%s", got)
		}

		errMsg, summary := split[0], split[1]
		if !strings.Contains(errMsg, "template syntax error") {
			t.Errorf("\nexp template syntax error\ngot %s", errMsg)
		}

		checkSummary(t, summary)
	})
}

// helpers

func newBenchmark() requester.Benchmark {
	return requester.Benchmark{
		Fail:     1,
		Success:  2,
		Length:   3,
		Duration: 4 * time.Second,
		Records: []requester.Record{
			{Time: time.Second},
			{Time: time.Second},
			{Time: time.Second},
		},
	}
}

func newConfig(tpl string) config.Global {
	urlURL, _ := url.ParseRequestURI("https://a.b.com")
	return config.Global{
		Request: config.Request{URL: urlURL},
		Runner:  config.Runner{Requests: -1},
		Output:  config.Output{Template: tpl},
	}
}

func checkSummary(t *testing.T, summary string) {
	t.Helper()

	expSummary := `
Endpoint           https://a.b.com
Requests           3/∞
Errors             1
Min response time  1000ms
Max response time  1000ms
Mean response time 1000ms
Total duration     4000ms
`[1:]

	if summary != expSummary {
		t.Errorf("\nexp summary:\n%q\ngot summary:\n%q", expSummary, summary)
	}
}
