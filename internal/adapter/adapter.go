package adapter

import (
	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/requester"
)

func RequesterConfig(cfg config.Global) requester.Config {
	return requester.Config(cfg.Runner)
}
