package repl

import (
	"fmt"
	"strconv"
)

type Command struct {
	Name     string
	Usage    string
	ArgCheck func(args []string) error
	Action   func(args []string) error
}

var (
	discoverDevicesREPLCmd = repl.Command{
		Name:  "discover",
		Usage: "discovers connected devices",
		Action: func(r *repl.REPL, args []string) error {
			err := perClerk.FindDevices()
			if err != nil {
				return err
			}

			return nil
		},
	}

	listDevicesREPLCmd = repl.Command{
		Name:  "list",
		Usage: "lists connected devices",
		Action: func(r *repl.REPL, args []string) error {
			_ = perClerk.ListDevices()

			return nil
		},
	}

	listDeviceAPIREPLCmd = repl.Command{
		Name:  "v-api",
		Usage: "shows API of a device - pass its ID",
		Action: func(r *repl.REPL, args []string) error {
			// We should add this to the REPL instead
			if len(args) < 1 {
				return fmt.Errorf("requires at least on argument")
			}

			// First and only argument should be the ID of the device we want to use
			targetID, err := strconv.ParseInt(args[0], 10, 0)
			if err != nil {
				return err
			}

			err = perClerk.ListDeviceAPI(uint8(targetID))
			if err != nil {
				fmt.Println("failed to view device API: ", err.Error())
			}

			return nil
		},
	}

	runDeviceAPIREPLCmd = repl.Command{
		Name:  "v-run",
		Usage: "runs a funcion of a device - pass its ID and function name",
		Action: func(r *repl.REPL, args []string) error {
			// We should add this to the REPL instead
			if len(args) < 3 {
				return fmt.Errorf("requires at least on argument")
			}

			// First and only argument should be the ID of the device we want to use
			targetID, err := strconv.ParseInt(args[0], 10, 0)
			if err != nil {
				return err
			}

			fnName := args[1]
			fnArgs := args[2:]

			err = perClerk.RunDeviceFunction(uint8(targetID), fnName, fnArgs)
			if err != nil {
				return err
			}

			return nil
		},
	}
)
