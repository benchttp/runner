package file

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

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
		// "aliases" is a native yaml field not expected to be marshaled,
		// yet Decode reports an error when decoder.KnownFields is set
		// to true: it erroneously expects a matching field in the destination
		// structure, so we discard these errors.
		if !strings.Contains(msg, "field aliases not found in type") {
			filtered.Errors = append(filtered.Errors, msg)
		}
	}

	if len(filtered.Errors) != 0 {
		return filtered
	}

	return nil
}

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
