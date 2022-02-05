package file

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"

	"gopkg.in/yaml.v3"
)

type extension string

const (
	extYML  extension = ".yml"
	extYAML extension = ".yaml"
	extJSON extension = ".json"
)

type configParser interface {
	parse(in []byte, dst interface{}) error
}

func newParser(ext extension) (configParser, error) {
	switch ext {
	case extYML, extYAML:
		return yamlParser{}, nil
	case extJSON:
		return jsonParser{}, nil
	default:
		return nil, errors.New("unsupported config format")
	}
}

type yamlParser struct{}

func (p yamlParser) parse(in []byte, dst interface{}) error {
	decoder := yaml.NewDecoder(bytes.NewReader(in))
	decoder.KnownFields(true)
	return p.handleError(decoder.Decode(dst))
}

// handleError handles a raw yaml decoder.Decode error, filters it,
// and return the resulting error.
func (p yamlParser) handleError(err error) error {
	// yaml.TypeError errors require special handling, other errors
	// (nil included) can be returned as is.
	var typeError *yaml.TypeError
	if !errors.As(err, &typeError) {
		return err
	}

	// filter out unwanted errors
	filtered := &yaml.TypeError{}
	for _, msg := range typeError.Errors {
		// With decoder.KnownFields set to true, Decode reports any field
		// that do not match the destination structure as a non-nil error.
		// It is a wanted behavior but prevents the usage of custom aliases.
		// To work around this we allow an exception for that rule with fields
		// starting with x- (inspired by docker compose api).
		if p.isCustomFieldError(msg) {
			continue
		}
		filtered.Errors = append(filtered.Errors, msg)
	}

	if len(filtered.Errors) != 0 {
		return filtered
	}

	return nil
}

func (p yamlParser) isCustomFieldError(raw string) bool {
	customFieldRgx := regexp.MustCompile(
		// raw output example:
		// 	line 9: field x-my-alias not found in type struct { ... }
		`^line \d+: field (x-[\w-]+) not found in type`,
	)
	return customFieldRgx.MatchString(raw)
}

// jsonParser implements configParser.
type jsonParser struct{}

func (jsonParser) parse(in []byte, dst interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(in))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

// unmarshaledConfig is a raw data model for runner config files.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
type unmarshaledConfig struct {
	Request struct {
		Method      *string           `yaml:"method" json:"method"`
		URL         *string           `yaml:"url" json:"url"`
		QueryParams map[string]string `yaml:"queryParams" json:"queryParams"`
		Timeout     *string           `yaml:"timeout" json:"timeout"`
	} `yaml:"request" json:"request"`

	RunnerOptions struct {
		Requests      *int    `yaml:"requests" json:"requests"`
		Concurrency   *int    `yaml:"concurrency" json:"concurrency"`
		GlobalTimeout *string `yaml:"globalTimeout" json:"globalTimeout"`
	} `yaml:"runnerOptions" json:"runnerOptions"`
}
