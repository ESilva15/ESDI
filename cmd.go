package main

import (
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "esdi",
	Short: "CLI for the dashDisplay project",
	Long:  `Allows the configuration and communication with the dashDisplay`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("out", "o", "", "output file")
	rootCmd.Flags().StringP("session", "s", "", "output session info to file")
}
