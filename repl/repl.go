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

func (r *REPL) RegisterCMD(newCMD Command) {
	// Add a verification to check that the command doesn't exist. Fatal if it does
	r.Commands[newCMD.Name] = newCMD
}

func (r *REPL) printPrompt() {
	_, _ = os.Stdout.Write([]byte(r.Cfg.PS1))
}

func (r *REPL) SetPS1(s string) {
	r.Cfg.PS1 = s
}

func parseArgs(l string) ([]string, error) {
	args := []string{}
	input := strings.TrimSpace(l)

	type ArgParserState uint8
	const (
		readingQuotedStr ArgParserState = iota
		readingNormalStr
		storeArg
	)

	curState := readingNormalStr

	curStr := []byte{}
	for k := 0; k < len(input); {
		switch curState {
		case readingNormalStr:
			switch input[k] {
			case '"':
				curState = readingQuotedStr
			case ' ':
				curState = storeArg
			default:
				curStr = append(curStr, input[k])
			}
			k++
		case readingQuotedStr:
			switch input[k] {
			case '"':
				if k+1 < len(input) {
					k++
				}
				curState = storeArg
			default:
				curStr = append(curStr, input[k])
			}
			k++
		case storeArg:
			args = append(args, string(curStr))
			curStr = nil
			curState = readingNormalStr
		}

		if k == len(input) {
			args = append(args, string(curStr))
			curStr = nil
		}
	}

	if curState == readingQuotedStr {
		return []string{}, fmt.Errorf("quoted string was not closed: %s", curStr)
	}

	return args, nil
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

		args, err := parseArgs(line)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}

		command, exists := r.Commands[args[0]]
		if exists {
			err = command.Action(r, args[1:])
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}
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
