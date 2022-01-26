package configparser

import (
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

type Parser interface {
	ParseConfig([]byte, interface{}) error
}

type yamlParser struct{}

func (yamlParser) ParseConfig(b []byte, dst interface{}) error {
	return yaml.Unmarshal(b, dst)
}

type jsonParser struct{}

func (jsonParser) ParseConfig(b []byte, dst interface{}) error {
	return json.Unmarshal(b, dst)
}

func newParser(ext extension) (Parser, error) {
	switch ext {
	case extYML, extYAML:
		return yamlParser{}, nil
	case extJSON:
		return jsonParser{}, nil
	default:
		return nil, errors.New("unsupported config format")
	}
}
