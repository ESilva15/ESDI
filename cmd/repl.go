package cmd

import (
	"esdi/peripheral"
	"esdi/repl"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func replCmdAction(cmd *cobra.Command, args []string) {
	r := repl.NewREPL(repl.REPLCfg{
		PS1: "\rESDI > ",
	})

	perClerk := peripheral.NewPeripheralDeviceClerk()

	discoverDevicesREPLCmd := repl.Command{
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

	listDevicesREPLCmd := repl.Command{
		Name:  "list",
		Usage: "lists connected devices",
		Action: func(r *repl.REPL, args []string) error {
			_ = perClerk.ListDevices()

			return nil
		},
	}

	listDeviceAPIREPLCmd := repl.Command{
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

	runDeviceAPIREPLCmd := repl.Command{
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

	r.RegisterCMD(discoverDevicesREPLCmd)
	r.RegisterCMD(listDevicesREPLCmd)
	r.RegisterCMD(listDeviceAPIREPLCmd)
	r.RegisterCMD(runDeviceAPIREPLCmd)

	r.Start()
	r.Close()
}

// removeLabelCmd represents the removeLabel command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "run the repl to interact with the display",
	Long:  ``,
	Run:   replCmdAction,
}

func init() {
	rootCmd.AddCommand(replCmd)
}
