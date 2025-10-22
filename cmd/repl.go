package cmd

import (
	"esdi/peripheral"
	"esdi/repl"

	"github.com/spf13/cobra"
)

func replCmdAction(cmd *cobra.Command, args []string) {
	r := repl.NewREPL(repl.REPLCfg{
		PS1: "\rESDI > ",
	})

	listDevicesREPLCmd := repl.Command{
		Name:  "ls",
		Usage: "lists the currently available devices",
		Action: func(r *repl.REPL, args []string) error {
			perClerk := peripheral.NewPeripheralDeviceClerk()
			err := perClerk.FindDevices()
			if err != nil {
				return err
			}

			return nil
		},
	}

	r.RegisterCMD(listDevicesREPLCmd)

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
