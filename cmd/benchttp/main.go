package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/benchttp/runner/config"
)

var errUsage = errors.New("usage")

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		if errors.Is(err, errUsage) {
			flag.Usage()
		}
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("%w: no command specified", errUsage)
	}

	args := os.Args[1:]

	switch sub := args[0]; sub {
	case "run":
		cmd := cmdRun{
			flagset:            flag.NewFlagSet("run", flag.ExitOnError),
			config:             config.Default(),
			defaultConfigFiles: defaultConfigFiles,
		}
		return cmd.execute(args)
	case "auth":
		cmd := cmdAuth{
			flagset: flag.NewFlagSet("auth", flag.ExitOnError),
		}
		return cmd.execute(args)
	default:
		return fmt.Errorf("%w: unknown command: %s", errUsage, sub)
	}
}
