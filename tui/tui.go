// Package tui
package tui

import (
	ctrls "esdi/tui/internal/controllers"
	"log/slog"
)

func Run(logger *slog.Logger) error {
	// // Call the main controller and let the program start!
	// return ctrl.Init(logger)

	// The app starts running here!
	//
	// Set the event capture for the global app itself here
	mc := ctrls.NewMainController(logger.With("[ctrl]", "main"))

	err := mc.Main()
	if err != nil {
		return err
	}

	return nil
}
