package main

import (
	"esdi/logger"
	"esdi/sources/iracing"

	"github.com/spf13/cobra"
)

func liveTelemetryCmdAction(cmd *cobra.Command, args []string) {
	log := logger.GetInstance()

	ddPort, _ := cmd.Flags().GetString("port")
	outputFile, _ := cmd.Flags().GetString("outputFile")

	log.Printf("Called `live`:\nPort: '%s'\nOutFile: '%s'\n", ddPort, outputFile)

	esdi, err := ESDIInit(ddPort, 115200)
	if err != nil {
		log.Fatalf("Failed to get Desktop Interface: %v", err)
	}

	irsdk, err := iracing.Init(nil)
	if err != nil {
		log.Fatalf("Failed to create iRacing interface: %v", err)
	}

	esdi.Source = &irsdk

	esdi.telemetry()
}

// removeLabelCmd represents the removeLabel command
var liveTelemetryCmd = &cobra.Command{
	Use:   "live",
	Short: "stream data directly from the game",
	Long:  ``,
	Run:   liveTelemetryCmdAction,
}

func init() {
	rootCmd.AddCommand(liveTelemetryCmd)

	// Declare the flags for this command
	liveTelemetryCmd.Flags().StringP("port", "p", "", "dashDisplay Port")

	// Mark the required ones
	liveTelemetryCmd.MarkFlagRequired("port")
}
