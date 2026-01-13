package ui

import (
	"esdi/tui/internal/models"

	"github.com/rivo/tview"
)

// Main controller events

type LogEv struct {
	Log string
}

type PrintLogEv struct {
	Log string
}

type RedrawEv struct{}

type ChangeFocusEv struct {
	Target tview.Primitive
}

// ----------------------------------------------------------------------------

// Layout events

type CreateWindowEv struct {
	Window models.Window
}

type WindowCreatedEv struct {
	Window models.Window
}

type ErrorCreateWindowEv struct {
	Error error
}

type ErrorFormParseEv struct {
	Error error
}
