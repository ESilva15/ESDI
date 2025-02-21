package main

import (
	"os"

	"esdi/logger"
	"esdi/sources/iracing"
	// "github.com/ESilva15/goirsdk"

	"github.com/spf13/cobra"
)

func offlineTelemetryCmdAction(cmd *cobra.Command, args []string) {
	log := logger.GetInstance()

	ddPort, _ := cmd.Flags().GetString("port")
	inFile, _ := cmd.Flags().GetString("in")
	outFile, _ := cmd.Flags().GetString("out")
	sessionFile, _ := cmd.Flags().GetString("session")

	log.Printf("Called `offline`:\nPort: '%s'\nSource: '%s'\nOutFile: '%s'\n", ddPort, inFile, outFile)

	esdi, err := ESDIInit(ddPort, 115200)
	if err != nil {
		log.Fatalf("Failed to get Desktop Interface: %v", err)
	}

	file, err := os.Open(inFile)
	if err != nil {
		log.Fatalf("Failed to open IBT file: %v", err)
	}

	irsdk, err := iracing.Init(file, outFile, sessionFile)
	if err != nil {
		log.Fatalf("Failed to create iRacing interface: %v", err)
	}
	// irsdk, err := goirsdk.Init(file, outFile, sessionFile)
	// if err != nil {
	// 	log.Fatalf("Failed to create irsdk instance: %v\n", err)
	// }

	esdi.irsdk = irsdk.SDK

	esdi.telemetry()
}

// removeLabelCmd represents the removeLabel command
var offlineTelemetryCmd = &cobra.Command{
	Use:   "offline",
	Short: "stream data from a file",
	Long:  ``,
	Run:   offlineTelemetryCmdAction,
}

func init() {
	rootCmd.AddCommand(offlineTelemetryCmd)

	// Declare the flags for this command
	offlineTelemetryCmd.Flags().StringP("port", "p", "", "dashDisplay Port")
	offlineTelemetryCmd.Flags().StringP("in", "i", "", "source file")

	// Mark the required ones
	offlineTelemetryCmd.MarkFlagRequired("port")
	offlineTelemetryCmd.MarkFlagRequired("in")
}
