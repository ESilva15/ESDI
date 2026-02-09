// Package cmd contains our cli
package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "esdi",
	Short: "CLI for the dashDisplay project",
	Long:  `Allows the configuration and communication with the dashDisplay`,
}

type loggerKey struct{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(log *slog.Logger) {
	ctx := context.WithValue(context.Background(), loggerKey{}, log)
	rootCmd.SetContext(ctx)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("out", "o", "", "output file")
	rootCmd.Flags().StringP("session", "s", "", "output session info to file")
}
