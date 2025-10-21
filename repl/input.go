package repl

import (
	"strings"
)

func parseInput(args string) []string {
	return strings.Split(args, " ")
}
