// Package tui
package tui

import (
	"esdi/tui/internal/controllers"
	"esdi/tui/internal/services"
	"log/slog"

	"github.com/rivo/tview"
)

type ControlPanel struct {
	*controllers.Controller

	DeviceController *controllers.DeviceController
}

func NewControlPanel(logger *slog.Logger) *ControlPanel {
	baseController := &controllers.Controller{
		Logger: logger,
		App:    tview.NewApplication(),
	}

	devService := services.NewCDashService(logger)
	telemService := services.NewTelemetryService().SetProvider("iRacing")
	if telemService == nil {
		panic("failed to create the telemetry service")
	}

	return &ControlPanel{
		Controller:       baseController,
		DeviceController: controllers.NewDeviceController(baseController, devService, telemService),
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
