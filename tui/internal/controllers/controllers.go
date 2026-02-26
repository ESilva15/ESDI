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
