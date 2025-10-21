package cmd

import (
	"esdi/esdi"
	"esdi/logger"

	"github.com/spf13/cobra"
)

func liveTelemetryCmdAction(cmd *cobra.Command, args []string) {
	log := logger.GetInstance()

	ddPort, _ := cmd.Flags().GetString("port")
	outputFile, _ := cmd.Flags().GetString("out")
	sessionFile, _ := cmd.Flags().GetString("session")

	log.Printf("Called `live`:\nPort: '%s'\nOutFile: '%s'\n", ddPort, outputFile)

	esdi.RunLiveTelemetry(ddPort, outputFile, sessionFile)
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
	// liveTelemetryCmd.Flags().StringP("out", "o", "", "output to an .ibt file")
	// liveTelemetryCmd.Flags().StringP("session", "s", "", "output session .yaml to file")

	// Mark the required ones
	liveTelemetryCmd.MarkFlagRequired("port")
}
