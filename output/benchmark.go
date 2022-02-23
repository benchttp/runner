package output

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/benchttp/runner/ansi"
	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/output/export"
	"github.com/benchttp/runner/requester"
)

// Benchmark represent a benchmark result as exported by the runner.
type Benchmark struct {
	Report   requester.Report
	Metadata struct {
		Config     config.Global
		FinishedAt time.Time
	}

	log func(v ...interface{})
}

// New returns an Benchmark initialized with rep and cfg.
func New(rep requester.Report, cfg config.Global) *Benchmark {
	outputLogger := newLogger(cfg.Output.Silent)
	return &Benchmark{
		Report: rep,
		Metadata: struct {
			Config     config.Global
			FinishedAt time.Time
		}{
			Config:     cfg,
			FinishedAt: time.Now(),
		},

		log: outputLogger.Println,
	}
}

// newLogger returns the logger to be used by Benchmark.
func newLogger(silent bool) *log.Logger {
	var w io.Writer = os.Stdout
	if silent {
		w = nopWriter{}
	}
	return log.New(w, ansi.Bold("→ "), 0)
}

// Export exports a Benchmark using the Strategies set in the attached
// config.Global. If any error occurs for a given Strategy, it does not
// block the other exports and returns an ExportError listing the errors.
func (bk Benchmark) Export() error {
	var ok bool
	var errs []error

	s := exportStrategy(bk.Metadata.Config.Output.Out)
	if s.is(Stdout) {
		bk.log(ansi.Bold("Summary"))
		export.Stdout(bk)
		ok = true
	}
	if s.is(JSONFile) {
		filename := genFilename()
		if err := export.JSONFile(filename, bk); err != nil {
			errs = append(errs, err)
		} else {
			bk.log(ansi.Bold("JSON generated"))
			fmt.Println(filename) // always print output filename
		}
		ok = true
	}
	if s.is(Benchttp) {
		if err := export.HTTP(bk); err != nil {
			errs = append(errs, err)
		} else {
			bk.log(ansi.Bold("Report sent to Benchttp"))
		}
		ok = true
	}

	if !ok {
		return ErrInvalidStrategy
	}
	if len(errs) != 0 {
		return &ExportError{Errors: errs}
	}
	return nil
}

// export.Interface implementation

var _ export.Interface = (*Benchmark)(nil)

// String returns a default summary of a Benchmark as a string.
func (bk Benchmark) String() string {
	var b strings.Builder

	if s, err := bk.applyTemplate(bk.Metadata.Config.Output.Template); err == nil {
		// template is non-empty and correctly executed,
		// returns its result instead of default summary.
		return s
	} else if errors.Is(err, errTemplateSyntax) {
		// template is non-empty but has syntax errors,
		// inform the user about it and fallback to default summary.
		b.WriteString(err.Error())
		b.WriteString("\nFalling back to default summary:\n")
	} // template is empty, use default summary.

	line := func(name string, value interface{}) string {
		const template = "%-18s %v\n"
		return fmt.Sprintf(template, name, value)
	}

	msString := func(d time.Duration) string {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	formatRequests := func(n, max int) string {
		maxString := strconv.Itoa(max)
		if maxString == "-1" {
			maxString = "∞"
		}
		return fmt.Sprintf("%d/%s", n, maxString)
	}

	var (
		cfg            = bk.Metadata.Config
		rep            = bk.Report
		min, max, mean = rep.Stats()
	)

	b.WriteString(line("Endpoint", cfg.Request.URL))
	b.WriteString(line("Requests", formatRequests(rep.Length, cfg.Runner.Requests)))
	b.WriteString(line("Errors", rep.Fail))
	b.WriteString(line("Min response time", msString(min)))
	b.WriteString(line("Max response time", msString(max)))
	b.WriteString(line("Mean response time", msString(mean)))
	b.WriteString(line("Test duration", msString(rep.Duration)))
	return b.String()
}

// applyTemplate applies Benchmark to a template using given pattern and returns
// the result as a string. If pattern == "", it returns errTemplateEmpty.
// If an error occurs parsing the pattern or executing the template,
// it returns errTemplateSyntax.
func (bk Benchmark) applyTemplate(pattern string) (string, error) {
	if pattern == "" {
		return "", errTemplateEmpty
	}

	t, err := template.New("template").Parse(bk.Metadata.Config.Output.Template)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errTemplateSyntax, err)
	}

	var b strings.Builder
	if err := t.Execute(&b, bk); err != nil {
		return "", fmt.Errorf("%w: %s", errTemplateSyntax, err)
	}

	return b.String(), nil
}

// HTTPRequest returns the *http.Request to be sent to Benchttp server.
// The Benchmark is encoded as gob in the request body.
func (bk Benchmark) HTTPRequest() (*http.Request, error) {
	// Encode request body as gob
	b, err := encodeGob(bk)
	if err != nil {
		return nil, err
	}

	// Create request
	r, err := http.NewRequest("POST", benchttpEndpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	return r, nil
}

// helpers

// encodeGob encodes the given Benchmark as gob-encoded bytes.
func encodeGob(bk Benchmark) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(bk); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// genFilename generates a JSON file name suffixed with a timestamp
// located in the working directory.
func genFilename() string {
	return fmt.Sprintf("./benchttp.report.%s.json", timestamp())
}

// timestamp returns the current time in format yy-mm-ddThh:mm:ssZhh:mm.
func timestamp() string {
	now := time.Now().UTC()
	y, m, d := now.Date()
	hh, mm, ss := now.Clock()
	return strings.ReplaceAll(
		fmt.Sprintf("%4d%2d%2d%2d%2d%2d", y, m, d, hh, mm, ss),
		" ", "0",
	)
}

type nopWriter struct{}

func (nopWriter) Write(b []byte) (int, error) { return 0, nil }
