// Package tui
package tui

import (
	"esdi/tui/internal/controllers"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
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
		Bus:    events.NewBus(),
		App:    tview.NewApplication(),
		Dom:    dom.NewDOM(),
	}

	deviceService := services.NewCDashService(logger)

	return &ControlPanel{
		Controller:       baseController,
		DeviceController: controllers.NewDeviceController(baseController, deviceService),
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
