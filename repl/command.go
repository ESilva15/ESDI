package repl

import (
	"fmt"
	"os"
)

var (
	helpCmd = Command{
		Name:  ".help",
		Usage: " type `.help` for help",
		Action: func(r *REPL, args []string) error {
			for _, cmd := range r.Commands {
				fmt.Printf("%-20s\t%s\n", cmd.Name, cmd.Usage)
			}

			return nil
		},
	}
	clearCmd = Command{
		Name:  ".clear",
		Usage: " type `.clear` to clear the screen",
		Action: func(r *REPL, args []string) error {
			_, _ = os.Stdout.Write([]byte("\033[2J\033[H"))
			return nil
		},
	}
	exitCmd = Command{
		Name:  ".exit",
		Usage: " type `.exit` to exit this REPL",
		Action: func(r *REPL, args []string) error {
			r.Stop = true
			return nil
		},
	}
)

type Command struct {
	Name   string
	Usage  string
	Action func(r *REPL, args []string) error
}
