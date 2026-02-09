package cmd

import (
	"esdi/tui"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func tuiCmdAction(cmd *cobra.Command, args []string) {
	logger, ok := cmd.Context().Value(loggerKey{}).(*slog.Logger)
	if !ok {
		fmt.Printf("Error loading logger from context: %s\n", logger)
		return
	}

	err := tui.Run(logger)
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
