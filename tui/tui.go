// Package tui
package tui

import (
	"log/slog"

	"esdi/services"
	"esdi/tui/internal/controllers"

	"github.com/rivo/tview"
)

type ControlPanel struct {
	*controllers.Controller

	DeviceController *controllers.DeviceController
}

func NewControlPanel(logger *slog.Logger) *ControlPanel {
	// NOTE: given the services should not be only for the TUI should this be here?
	baseController := &controllers.Controller{
		Logger: logger,
		App:    tview.NewApplication(),
	}

	orchestrator, err := services.NewOrchestrator(logger)
	if err != nil {
		// TODO: no panic here
		panic("failed to create services orchestrator")
	}

	return &ControlPanel{
		Controller:       baseController,
		DeviceController: controllers.NewDeviceController(baseController, orchestrator),
	}
}

func (cp *ControlPanel) Run() error {
	if err := cp.DeviceController.Main(); err != nil {
		return err
	}

	return cp.App.Run()
}

func Run(logger *slog.Logger) error {
	progController := NewControlPanel(logger)
	return progController.Run()
}
