package cmd

import (
	"esdi/tui"
	"fmt"

	"github.com/spf13/cobra"
)

func tuiCmdAction(cmd *cobra.Command, args []string) {
	err := tui.Init()
	if err != nil {
		fmt.Printf("error running TUI: %s\n", err.Error())
	}
}

// removeLabelCmd represents the removeLabel command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "runs an interactive application",
	Long:  ``,
	Run:   tuiCmdAction,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
