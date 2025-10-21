// Package repl Basic UI for the user
package repl

import (
	"bufio"
	"fmt"
	"os"
)

var DefaultCfg = REPLCfg{
	PS1: "\r> ",
}

type REPLCfg struct {
	PS1 string
}

type REPL struct {
	Cfg      REPLCfg
	Commands map[string]Command
}

func NewREPL(cfg REPLCfg) *REPL {
	return &REPL{
		Cfg: cfg,
		Commands: map[string]Command{
			helpCmd.Name:  helpCmd,
			clearCmd.Name: clearCmd,
			exitCmd.Name:  exitCmd,
		},
	}
}

func (r *REPL) printPrompt() {
	_, _ = os.Stdout.Write([]byte(r.Cfg.PS1))
}

func (r *REPL) Start() {
	reader := bufio.NewScanner(os.Stdin)
	r.printPrompt()
	for reader.Scan() {
		input := reader.Text()
		args := parseInput(input)

		command, exists := r.Commands[args[0]]
		if exists {
			_ = command.Action(r, args[1:])
		} else {
			fmt.Printf("No such command `%s`\n", args[0])
		}

		r.printPrompt()
	}
}
