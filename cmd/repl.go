package cmd

import (
	"esdi/repl"

	"github.com/spf13/cobra"
)

func replCmdAction(cmd *cobra.Command, args []string) {
	repl := repl.NewREPL(repl.REPLCfg{
		PS1: "\rESDI > ",
	})
	repl.Start()
	repl.Close()
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
