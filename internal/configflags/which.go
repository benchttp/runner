package configflags

import (
	"flag"

	"github.com/benchttp/runner/config"
)

// Which returns a slice of all config fields set via the CLI.
func Which() []string {
	var fields []string
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if name := f.Name; config.IsField(name) {
			fields = append(fields, name)
		}
	})
	return fields
}
