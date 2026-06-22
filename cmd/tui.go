package cmd

import (
	"fmt"
	"log/slog"

	"esdi/tui"

	"github.com/spf13/cobra"
)

func tuiCmdAction(cmd *cobra.Command, args []string) {
	err := tui.Run(slog.Default())
	if err != nil {
		fmt.Printf("Error running TUI: %s\n", err.Error())
		return
	}
}

// removeLabelCmd represents the removeLabel command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "runs a TUI",
	Long:  ``,
	Run:   tuiCmdAction,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
