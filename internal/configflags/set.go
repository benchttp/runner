package configflags

import (
	"flag"
	"net/http"
	"net/url"

	"github.com/benchttp/runner/config"
)

// Set reads arguments provided to flagset as config.Fields and binds
// their value to the appropriate fields of given *config.Global.
func Set(flagset *flag.FlagSet, cfg *config.Global) {
	// avoid nil pointer dereferences
	if cfg.Request.URL == nil {
		cfg.Request.URL = &url.URL{}
	}
	if cfg.Request.Header == nil {
		cfg.Request.Header = http.Header{}
	}

	// request url
	flagset.Var(urlValue{url: cfg.Request.URL},
		config.FieldURL,
		config.FieldsDesc[config.FieldURL],
	)
	// request method
	flagset.StringVar(&cfg.Request.Method,
		config.FieldMethod,
		cfg.Request.Method,
		config.FieldsDesc[config.FieldMethod],
	)
	// request header
	flagset.Var(headerValue{header: &cfg.Request.Header},
		config.FieldHeader,
		config.FieldsDesc[config.FieldHeader],
	)
	// request body
	flagset.Var(bodyValue{body: &cfg.Request.Body},
		config.FieldBody,
		config.FieldsDesc[config.FieldBody],
	)
	// requests number
	flagset.IntVar(&cfg.Runner.Requests,
		config.FieldRequests,
		cfg.Runner.Requests,
		config.FieldsDesc[config.FieldRequests],
	)

	// concurrency
	flagset.IntVar(&cfg.Runner.Concurrency,
		config.FieldConcurrency,
		cfg.Runner.Concurrency,
		config.FieldsDesc[config.FieldConcurrency],
	)
	// non-conurrent requests interval
	flagset.DurationVar(&cfg.Runner.Interval,
		config.FieldInterval,
		cfg.Runner.Interval,
		config.FieldsDesc[config.FieldInterval],
	)
	// request timeout
	flagset.DurationVar(&cfg.Runner.RequestTimeout,
		config.FieldRequestTimeout,
		cfg.Runner.RequestTimeout,
		config.FieldsDesc[config.FieldRequestTimeout],
	)
	// global timeout
	flagset.DurationVar(&cfg.Runner.GlobalTimeout,
		config.FieldGlobalTimeout,
		cfg.Runner.GlobalTimeout,
		config.FieldsDesc[config.FieldGlobalTimeout],
	)

	// output strategies
	flagset.Var(outValue{out: &cfg.Output.Out},
		config.FieldOut,
		config.FieldsDesc[config.FieldOut],
	)
	// silent mode
	flagset.BoolVar(&cfg.Output.Silent,
		config.FieldSilent,
		cfg.Output.Silent,
		config.FieldsDesc[config.FieldSilent],
	)
	// output template
	flagset.StringVar(&cfg.Output.Template,
		config.FieldTemplate,
		cfg.Output.Template,
		config.FieldsDesc[config.FieldTemplate],
	)
}
