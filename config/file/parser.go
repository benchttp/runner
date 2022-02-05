package file

import (
	"bytes"
	"encoding/json"
	"errors"

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

type yamlParser struct{}

func (yamlParser) parse(in []byte, dst interface{}) error {
	decoder := yaml.NewDecoder(bytes.NewReader(in))
	decoder.KnownFields(true)
	return decoder.Decode(dst)
}

type jsonParser struct{}

func (jsonParser) parse(in []byte, dst interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(in))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
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
