package main

import (
	"flag"
	"fmt"
)

// cmdAuth handles subcommand "benchttp auth [options]".
type cmdAuth struct {
	flagset *flag.FlagSet
}

func (cmdAuth) execute(_ []string) error {
	fmt.Println("Benchttp authentication")
	return nil
}
