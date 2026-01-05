package tui

import (
	"esdi/peripheral"
	"fmt"
	"strconv"
	"strings"
)

type Command struct {
	Name     string
	Usage    string
	ArgCheck func(args []string) error
	Action   func(args []string) (string, error)
}

func (c *Command) Run(args []string) (string, error) {
	if c.ArgCheck != nil {
		err := c.ArgCheck(args)
		if err != nil {
			return "", err
		}
	}

	return c.Action(args)
}

type CommandManager struct {
	Commands map[string]Command
}

func NewCommandManager() *CommandManager {
	return &CommandManager{
		Commands: make(map[string]Command),
	}
}

func (cm *CommandManager) RegisterCommand(c Command) {
	cm.Commands[c.Name] = c
}

func (cm *CommandManager) RunCommand(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("require at least one argument: the command")
	}

	command, exists := cm.Commands[args[0]]
	if !exists {
		return "", fmt.Errorf("no such command `%s`", args[0])
	}

	return command.Run(args[1:])
}

func getCommands() *CommandManager {
	perClerk := peripheral.NewPeripheralDeviceClerk()
	cm := NewCommandManager()

	discoverDevicesREPLCmd := Command{
		Name:  "discover",
		Usage: "discovers connected devices",
		Action: func(args []string) (string, error) {
			output, err := perClerk.FindDevices()
			if err != nil {
				return "", err
			}

			return output, nil
		},
	}
	cm.RegisterCommand(discoverDevicesREPLCmd)

	listDevicesREPLCmd := Command{
		Name:  "list",
		Usage: "lists connected devices",
		Action: func(args []string) (string, error) {
			output, err := perClerk.ListDevices()
			if err != nil {
				return "", err
			}

			return output, nil
		},
	}
	cm.RegisterCommand(listDevicesREPLCmd)

	listDeviceAPIREPLCmd := Command{
		Name:  "v-api",
		Usage: "shows API of a device - pass its ID",
		ArgCheck: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires at least on argument")
			}

			return nil
		},
		Action: func(args []string) (string, error) {
			// First and only argument should be the ID of the device we want to use
			targetID, err := strconv.ParseInt(args[0], 10, 0)
			if err != nil {
				return "", err
			}

			output, err := perClerk.ListDeviceAPI(uint8(targetID))
			if err != nil {
				fmt.Println("failed to view device API: ", err.Error())
			}

			return output, nil
		},
	}
	cm.RegisterCommand(listDeviceAPIREPLCmd)

	runDeviceAPIREPLCmd := Command{
		Name:  "v-run",
		Usage: "runs a funcion of a device - pass its ID and function name",
		ArgCheck: func(args []string) error {
			// We should add this to the REPL instead
			if len(args) < 3 {
				return fmt.Errorf("requires at least on argument")
			}
			return nil
		},
		Action: func(args []string) (string, error) {
			// First and only argument should be the ID of the device we want to use
			targetID, err := strconv.ParseInt(args[0], 10, 0)
			if err != nil {
				return "", err
			}

			fnName := args[1]
			fnArgs := args[2:]

			output, err := perClerk.RunDeviceFunction(uint8(targetID), fnName, fnArgs)
			if err != nil {
				return "", err
			}

			return output, nil
		},
	}
	cm.RegisterCommand(runDeviceAPIREPLCmd)

	helpCmd := Command{
		Name:  ".help",
		Usage: " type `.help` for help",
		Action: func(args []string) (string, error) {
			var s strings.Builder
			for _, cmd := range cm.Commands {
				s.WriteString(fmt.Sprintf("%-20s\t%s\n", cmd.Name, cmd.Usage))
			}

			return s.String(), nil
		},
	}
	cm.RegisterCommand(helpCmd)

	return cm
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
