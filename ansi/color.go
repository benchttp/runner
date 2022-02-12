package ansi

import "fmt"

type color string

const (
	green  color = "\033[1;32m"
	yellow color = "\033[1;33m"
	red    color = "\033[1;31m"
	grey   color = "\033[1;30m"
	reset  color = "\033[0m"
)

func colorize(in string, c color) string {
	return fmt.Sprintf("%s%s%s", c, in, reset)
}

func Green(in string) string {
	return colorize(in, green)
}

func Yellow(in string) string {
	return colorize(in, yellow)
}

func Red(in string) string {
	return colorize(in, red)
}

func Grey(in string) string {
	return colorize(in, grey)
}
