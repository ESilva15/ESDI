// Package repl Basic UI for the user
package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var DefaultCfg = REPLCfg{
	PS1: "\r> ",
}

type REPLCfg struct {
	PS1 string
}

type REPL struct {
	Stop     bool
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

func (r *REPL) Close() {
	// Do whatever cleanup we have to do
}

func (r *REPL) RegisterCMD() {
	// Code to register a new command
}

func (r *REPL) printPrompt() {
	_, _ = os.Stdout.Write([]byte(r.Cfg.PS1))
}

func (r *REPL) mainLoop() {
	reader := bufio.NewReader(os.Stdin)
	for {
		r.printPrompt()
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error: ", err)
			continue
		}

		args := parseInput(strings.TrimSpace(line))
		if args[0] == "" {
			continue
		}

		command, exists := r.Commands[args[0]]
		if exists {
			_ = command.Action(r, args[1:])
		} else {
			fmt.Printf("No such command `%s`\n", args[0])
		}

		if r.Stop {
			break
		}
	}
}

func (r *REPL) Start() {
	// Setup handlers and whatnot
	r.mainLoop()
}
