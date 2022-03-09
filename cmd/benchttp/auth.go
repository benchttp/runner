package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/benchttp/runner/ansi"
	"github.com/benchttp/runner/internal/auth"
)

// cmdAuth handles subcommand "benchttp auth [options]".
type cmdAuth struct {
	flagset *flag.FlagSet
}

// ensure cmdAuth implements command
var _ command = (*cmdAuth)(nil)

func (cmd cmdAuth) execute(args []string) error {
	if len(args) != 2 {
		cmd.flagset.Usage()
		return errUsage
	}

	switch sub := args[1]; sub {
	case "login":
		return cmd.login()
	case "logout":
		return cmd.logout()
	default:
		return fmt.Errorf("%w: unknown subcommand: %s", errUsage, sub)
	}
}

const (
	tokenURL  = "https://www.benchttp.app/login" // nolint:gosec // no creds
	tokenDir  = ".config/benchttp"               // nolint:gosec // no creds
	tokenName = "token.txt"
)

func (cmd cmdAuth) login() error {
	token, err := promptf("Visit %s and paste the token:\n", tokenURL)
	if err != nil {
		return err
	}

	tokenPath, err := cmd.tokenPath()
	if err != nil {
		return err
	}

	if err := auth.SaveToken(tokenPath, token); err != nil {
		return err
	}

	fmt.Printf("%sSuccessfully logged in.\n", ansi.Erase(1))
	return nil
}

func (cmd cmdAuth) logout() error {
	tokenPath, err := cmd.tokenPath()
	if err != nil {
		return err
	}

	if err := auth.SaveToken(tokenPath, ""); err != nil {
		return err
	}

	fmt.Println("Successfully logged out.")
	return nil
}

func (cmd cmdAuth) tokenPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		// TODO: handle error
		return "", err
	}

	dir := filepath.Join(home, tokenDir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		// TODO: handle error
		return "", err
	}

	return filepath.Join(dir, tokenName), nil
}
