package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
)

// errUsage reports an incorrect usage of the benchttp command.
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

	var cmd command
	args := os.Args[1:]

	switch sub := args[0]; sub {
	case "run":
		cmd = &cmdRun{flagset: flag.NewFlagSet("run", flag.ExitOnError)}
	case "auth":
		cmd = &cmdAuth{flagset: flag.NewFlagSet("auth", flag.ExitOnError)}
	case "version":
		cmd = &cmdVersion{}
	default:
		return fmt.Errorf("%w: unknown command: %s", errUsage, sub)
	}

	return cmd.execute(args)
}

// command is the interface that all benchttp subcommands must implement.
type command interface {
	execute(args []string) error
}

func promptf(message string, v ...interface{}) (string, error) {
	fmt.Printf(message, v...)
	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}
