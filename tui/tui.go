// Package tui
package tui

import (
	"esdi/tui/internal/controllers"
	"log/slog"
)

func Run(logger *slog.Logger) error {
	progController := controllers.NewControlPanel(logger)
	return progController.Run()
}
