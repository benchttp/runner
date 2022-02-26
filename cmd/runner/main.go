package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/benchttp/runner/config"
	configfile "github.com/benchttp/runner/config/file"
	"github.com/benchttp/runner/output"
	"github.com/benchttp/runner/requester"
)

var (
	configFile string

	cliConfig config.Global
)

var defaultConfigFiles = []string{
	"./.benchttp.yml",
	"./.benchttp.yaml",
	"./.benchttp.json",
}

func parseArgs() {
	cliConfig.Request.URL = &url.URL{}
	cliConfig.Request.Header = http.Header{}

	// config file path
	flag.StringVar(&configFile, "configFile", configfile.Find(defaultConfigFiles), "Config file path")

	// request url
	flag.Var(&urlValue{url: cliConfig.Request.URL}, config.FieldURL, "HTTP request url")

	// request method
	flag.StringVar(&cliConfig.Request.Method, config.FieldMethod, "", "HTTP request method")
	// request header
	flag.Var(headerValue{header: &cliConfig.Request.Header}, config.FieldHeader, "HTTP request header")
	// request body
	flag.Var(bodyValue{body: &cliConfig.Request.Body}, config.FieldBody, "HTTP request body")

	// requests number
	flag.IntVar(&cliConfig.Runner.Requests, config.FieldRequests, 0, "Number of requests to run, use duration as exit condition if omitted")
	// concurrency
	flag.IntVar(&cliConfig.Runner.Concurrency, config.FieldConcurrency, 0, "Number of connections to run concurrently")
	// non-conurrent requests interval
	flag.DurationVar(&cliConfig.Runner.Interval, "interval", 0, "Minimum duration between two non concurrent requests")
	// request timeout
	flag.DurationVar(&cliConfig.Runner.RequestTimeout, config.FieldRequestTimeout, 0, "Timeout for each HTTP request")
	// global timeout
	flag.DurationVar(&cliConfig.Runner.GlobalTimeout, config.FieldGlobalTimeout, 0, "Max duration of test")

	// output strategies
	flag.Var(outValue{out: &cliConfig.Output.Out}, config.FieldOut, "Output destination (benchttp,json,stdout)")
	// silent mode
	flag.BoolVar(&cliConfig.Output.Silent, config.FieldSilent, false, "Silent mode (no write to stdout)")
	// output template
	flag.StringVar(&cliConfig.Output.Template, "template", "", "Custom stdout output template")

	flag.Parse()
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	parseArgs()

	fmt.Println(cliConfig.Output.Out)

	cfg, err := parseConfig()
	if err != nil {
		return err
	}

	req, err := cfg.Request.Value()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go listenOSInterrupt(cancel)

	rep, err := requester.New(requesterConfig(cfg)).Run(ctx, req)
	if err != nil {
		if errors.Is(err, requester.ErrCanceled) {
			if err := handleRunInterrupt(); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return output.New(rep, cfg).Export()
}

// parseConfig returns a config.Config initialized with config file
// options if found, overridden with CLI options.
func parseConfig() (cfg config.Global, err error) {
	fileConfig, err := configfile.Parse(configFile)
	if err != nil && !errors.Is(err, configfile.ErrFileNotFound) {
		// config file is not mandatory, other errors are critical
		return
	}

	mergedConfig := fileConfig.Override(cliConfig, configFlags()...)

	return mergedConfig, mergedConfig.Validate()
}

// configFlags returns a slice of all config fields set via the CLI.
func configFlags() []string {
	var fields []string
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if name := f.Name; config.IsField(name) {
			fields = append(fields, name)
		}
	})
	return fields
}

// requesterConfig returns a requester.Config generated from cfg.
func requesterConfig(cfg config.Global) requester.Config {
	return requester.Config{
		Requests:       cfg.Runner.Requests,
		Concurrency:    cfg.Runner.Concurrency,
		Interval:       cfg.Runner.Interval,
		RequestTimeout: cfg.Runner.RequestTimeout,
		GlobalTimeout:  cfg.Runner.GlobalTimeout,
		Silent:         cfg.Output.Silent,
	}
}

// headerValue implements flag.Value
type headerValue struct {
	header *http.Header
}

// String returns a string representation of the referenced header.
func (v headerValue) String() string {
	return fmt.Sprint(v.header)
}

// Set reads input string in format "key:value" and appends value
// to the key's values of the referenced header.
func (v headerValue) Set(raw string) error {
	keyval := strings.SplitN(raw, ":", 2)
	if len(keyval) != 2 {
		return errors.New(`expect format "<key>:<value>"`)
	}
	key, val := keyval[0], keyval[1]
	(*v.header)[key] = append((*v.header)[key], val)
	return nil
}

// bodyValue implements flag.Value
type bodyValue struct {
	body *config.Body
}

// String returns a string representation of the referenced body.
func (v bodyValue) String() string {
	return fmt.Sprint(v.body)
}

// Set reads input string in format "type:content" and sets
// the referenced body accordingly.
//
// If type is "raw", content is the data as a string.
//	"raw:{\"key\":\"value\"}" // escaped JSON
//	"raw:text" // plain text
// If type is "file", content is the path to the file holding the data.
//	"file:./path/to/file.txt"
//
// Note: only type "raw" is supported at the moment.
func (v bodyValue) Set(raw string) error {
	errFormat := fmt.Errorf(`expect format "<type>:<content>", got "%s"`, raw)

	if raw == "" {
		return errFormat
	}

	split := strings.SplitN(raw, ":", 2)
	if len(split) != 2 {
		return errFormat
	}

	btype, bcontent := split[0], split[1]
	if bcontent == "" {
		return errFormat
	}

	switch btype {
	case "raw":
		*v.body = config.NewBody(btype, bcontent)
	// case "file":
	// 	// TODO
	default:
		return fmt.Errorf(`unsupported type: %s (only "raw" accepted`, btype)
	}
	return nil
}

// outValue implements flag.Value
type outValue struct {
	out *[]config.OutputStrategy
}

// String returns a string representation of outValue.out.
func (v outValue) String() string {
	return fmt.Sprint(v.out)
}

// Set reads input string as comma-separated values and appends the values
// to the key's values of the referenced header.
func (v outValue) Set(in string) error {
	values := strings.Split(in, ",")
	if len(values) < 1 {
		return errors.New(`expect comma-separated values`)
	}
	for _, value := range values {
		*v.out = append(*v.out, config.OutputStrategy(value))
	}
	return nil
}

// urlValue implements flag.Value
type urlValue struct {
	url *url.URL
}

// String returns a string representation of urlValue.url.
func (v urlValue) String() string {
	if v.url == nil {
		return ""
	}
	return v.url.String()
}

// Set parses input string as a URL and sets the referenced URL accordingly.
func (v urlValue) Set(in string) error {
	urlURL, err := url.ParseRequestURI(in)
	if err != nil {
		return fmt.Errorf(`invalid url: "%s"`, in)
	}
	*v.url = *urlURL
	return nil
}

// listenOSInterrupt listens for OS interrupt signals and calls callback.
// It should be called in a separate goroutine from main as it blocks
// the execution until the OS interrupt signal is received.
func listenOSInterrupt(callback func()) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	<-sigC
	callback()
}

// handleRunInterrupt handles the case when the runner is interrupted.
func handleRunInterrupt() error {
	reader := bufio.NewReader(os.Stdin)
	// TODO: list output strategies
	// TODO: do not prompt if strategy is stdout only
	// TODO: add config option "output.generateOnCancel" and remove prompt?
	fmt.Printf("\nBenchmark interrupted, generate output anyway? (yes/no): ")
	line, _, err := reader.ReadLine()
	if err != nil {
		return err
	}
	if string(line) != "yes" {
		return errors.New("benchmark interrupted without output")
	}
	return nil
}
