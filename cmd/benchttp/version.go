package main

import "fmt"

type cmdVersion struct {
	version string
}

// ensure cmdVersion implements command
var _ command = (*cmdVersion)(nil)

func (cmd cmdVersion) execute([]string) error {
	fmt.Println("benchttp", cmd.version)
	return nil
}
