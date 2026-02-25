// Package controllers defines the controller stuff for our TUI
package controllers

import (
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"log/slog"

	"github.com/rivo/tview"
)

type Controller struct {
	Logger *slog.Logger
	Bus    *events.Bus
	App    *tview.Application
	Dom    *dom.DOM
}

type ControlPanel struct {
	*Controller

	DeviceController *DeviceController
}

func NewControlPanel(logger *slog.Logger) *ControlPanel {
	baseController := &Controller{
		Logger: logger,
		Bus:    events.NewBus(),
		App:    tview.NewApplication(),
		Dom:    dom.NewDOM(),
	}

	return &ControlPanel{
		Controller:       baseController,
		DeviceController: NewDeviceController(baseController),
	}
}

func (cp *ControlPanel) Run() error {
	if err := cp.DeviceController.Main(); err != nil {
		return err
	}

	return cp.App.Run()
}
