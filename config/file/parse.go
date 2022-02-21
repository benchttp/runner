package file

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/benchttp/runner/config"
)

// unmarshaledConfig is a raw data model for runner config files.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
type unmarshaledConfig struct {
	Extends *string `yaml:"extends" json:"extends"`

	Request struct {
		Method      *string             `yaml:"method" json:"method"`
		URL         *string             `yaml:"url" json:"url"`
		QueryParams map[string]string   `yaml:"queryParams" json:"queryParams"`
		Header      map[string][]string `yaml:"header" json:"header"`
		Body        *struct {
			Type    string `yaml:"type" json:"type"`
			Content string `yaml:"content" json:"content"`
		} `yaml:"body" json:"body"`
	} `yaml:"request" json:"request"`

	Runner struct {
		Requests       *int    `yaml:"requests" json:"requests"`
		Concurrency    *int    `yaml:"concurrency" json:"concurrency"`
		Interval       *string `yaml:"interval" json:"interval"`
		RequestTimeout *string `yaml:"requestTimeout" json:"requestTimeout"`
		GlobalTimeout  *string `yaml:"globalTimeout" json:"globalTimeout"`
	} `yaml:"runner" json:"runner"`

	Output struct {
		Out      *[]string `yaml:"out" json:"out"`
		Silent   *bool     `yaml:"silent" json:"silent"`
		Template *string   `yaml:"template" json:"template"`
	} `yaml:"output" json:"output"`
}

// Parse parses a benchttp runner config file into a config.Global
// and returns it or the first non-nil error occurring in the process,
// which can be any of the values declared in the package.
func Parse(cfgpath string) (cfg config.Global, err error) {
	uconfs, err := parseFileRecursive(cfgpath, []unmarshaledConfig{}, map[string]bool{})
	if err != nil {
		return cfg, err
	}
	return parseAndMergeConfigs(uconfs)
}

// parseFileRecursive parses a config file and its parent found from key
// "extends" recursively until the root config file is reached.
// It returns the list of all parsed configs or the first non-nil error
// occurring in the process.
func parseFileRecursive(
	cfgpath string,
	uconfs []unmarshaledConfig,
	seen map[string]bool,
) ([]unmarshaledConfig, error) {
	// parse current config file
	uconf, err := parseFile(cfgpath)
	if err != nil {
		return uconfs, err
	}

	// root config reached, append it and return the result
	if uconf.Extends == nil {
		uconfs = append(uconfs, uconf)
		return uconfs, nil
	}

	// avoid infinite recursion caused by circular reference
	if _, exists := seen[*uconf.Extends]; exists {
		return uconfs, ErrCircularExtends
	}

	// record file, append config
	seen[*uconf.Extends] = true
	uconfs = append(uconfs, uconf)

	// resolve extended config path
	parentPath := filepath.Join(filepath.Dir(cfgpath), *uconf.Extends)

	// parse parent config file
	return parseFileRecursive(parentPath, uconfs, seen)
}

// parseFile parses a single config file and returns the result as an
// unmarshaledConfig and an appropriate error predeclared in the package.
func parseFile(cfgpath string) (uconf unmarshaledConfig, err error) {
	b, err := os.ReadFile(cfgpath)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return uconf, errWithDetails(ErrFileNotFound, cfgpath)
	default:
		return uconf, errWithDetails(ErrFileRead, cfgpath, err)
	}

	ext := extension(filepath.Ext(cfgpath))
	parser, err := newParser(ext)
	if err != nil {
		return uconf, errWithDetails(ErrFileExt, ext, err)
	}

	if err = parser.parse(b, &uconf); err != nil {
		return uconf, errWithDetails(ErrParse, cfgpath, err)
	}

	return uconf, nil
}

// parsedConfig holds a parsed config.Global and the list of its set fields.
type parsedConfig struct {
	value  config.Global
	fields []string
}

// parseAndMergeConfigs iterates backwards over uconfs, parsing them
// as config.Global and merging them into a single one.
// It returns the merged result or the first non-nil error occurring in the
// process.
func parseAndMergeConfigs(uconfs []unmarshaledConfig) (cfg config.Global, err error) {
	if len(uconfs) == 0 { // supposedly catched upstream, should not occur
		return cfg, errors.New(
			"an unacceptable error occurred parsing the config file, " +
				"please visit https://github.com/benchttp/runner/issues/new " +
				"and insult us properly",
		)
	}

	cfg = config.Default()

	for i := len(uconfs) - 1; i >= 0; i-- {
		rawCfg := uconfs[i]
		parsedCfg, err := parseRawConfig(rawCfg)
		if err != nil {
			return cfg, errWithDetails(ErrParse, "", err)
		}
		cfg = cfg.Override(parsedCfg.value, parsedCfg.fields...)
	}

	return cfg, nil
}

// parseRawConfig parses an input raw config as a config.Global and returns
// a parsedConfig or the first non-nil error occurring in the process.
func parseRawConfig(uconf unmarshaledConfig) (parsedConfig, error) { //nolint:gocognit // acceptable complexity for a parsing func
	const numField = 12 // should match the number of config Fields (not critical)

	cfg := config.Global{}
	fields := make([]string, 0, numField)

	appendField := func(field string) {
		fields = append(fields, field)
	}

	if method := uconf.Request.Method; method != nil {
		cfg.Request.Method = *method
		appendField(config.FieldMethod)
	}

	if rawURL := uconf.Request.URL; rawURL != nil {
		parsedURL, err := parseAndBuildURL(*uconf.Request.URL, uconf.Request.QueryParams)
		if err != nil {
			return parsedConfig{}, err
		}
		cfg.Request.URL = parsedURL
		appendField(config.FieldURL)
	}

	if header := uconf.Request.Header; header != nil {
		httpHeader := http.Header{}
		for key, val := range header {
			httpHeader[key] = val
		}
		cfg.Request.Header = httpHeader
		appendField(config.FieldHeader)
	}

	if body := uconf.Request.Body; body != nil {
		cfg.Request.Body = config.Body{
			Type:    body.Type,
			Content: []byte(body.Content),
		}
		fields = append(fields, config.FieldBody)
	}

	if requests := uconf.Runner.Requests; requests != nil {
		cfg.Runner.Requests = *requests
		appendField(config.FieldRequests)
	}

	if concurrency := uconf.Runner.Concurrency; concurrency != nil {
		cfg.Runner.Concurrency = *concurrency
		appendField(config.FieldConcurrency)
	}

	if interval := uconf.Runner.Interval; interval != nil {
		parsedInterval, err := parseOptionalDuration(*interval)
		if err != nil {
			return parsedConfig{}, err
		}
		cfg.Runner.Interval = parsedInterval
		appendField(config.FieldInterval)
	}

	if requestTimeout := uconf.Runner.RequestTimeout; requestTimeout != nil {
		parsedTimeout, err := parseOptionalDuration(*requestTimeout)
		if err != nil {
			return parsedConfig{}, err
		}
		cfg.Runner.RequestTimeout = parsedTimeout
		appendField(config.FieldRequestTimeout)
	}

	if globalTimeout := uconf.Runner.GlobalTimeout; globalTimeout != nil {
		parsedGlobalTimeout, err := parseOptionalDuration(*globalTimeout)
		if err != nil {
			return parsedConfig{}, err
		}
		cfg.Runner.GlobalTimeout = parsedGlobalTimeout
		appendField(config.FieldGlobalTimeout)
	}

	if out := uconf.Output.Out; out != nil {
		for _, o := range *out {
			cfg.Output.Out = append(cfg.Output.Out, config.OutputStrategy(o))
		}
		appendField(config.FieldOut)
	}

	if silent := uconf.Output.Silent; silent != nil {
		cfg.Output.Silent = *silent
		appendField(config.FieldSilent)
	}

	if template := uconf.Output.Template; template != nil {
		cfg.Output.Template = *template
		appendField(config.FieldTemplate)
	}

	return parsedConfig{
		value:  cfg,
		fields: fields,
	}, nil
}

// parseAndBuildURL parses a raw string as a *url.URL and adds any extra
// query parameters. It returns the first non-nil error occurring in the
// process.
func parseAndBuildURL(raw string, qp map[string]string) (*url.URL, error) {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return nil, err
	}

	// retrieve url query, add extra params, re-attach to url
	if qp != nil {
		q := u.Query()
		for k, v := range qp {
			q.Add(k, v)
		}
		u.RawQuery = q.Encode()
	}

	return u, nil
}

// parseOptionalDuration parses the raw string as a time.Duration
// and returns the parsed value or a non-nil error.
// Contrary to time.ParseDuration, it does not return an error
// if raw == "".
func parseOptionalDuration(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	return time.ParseDuration(raw)
}
