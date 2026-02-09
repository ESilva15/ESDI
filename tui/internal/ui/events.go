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

type DestroyWindowEv struct {
	ID int16
}

type LayoutRegisterWindowEv struct {
	Window models.Window
}

type WindowCreatedEv struct {
	ID    int16
	Title string
}

type WindowDestroyedEv struct {
	ID int16
}

type ErrorCreateWindowEv struct {
	Error error
}

type ErrorFormParseEv struct {
	Error error
}

// Layout events --------------------------------------------------------------

// Device events

type FindCDashDisplay struct{}

// Device events --------------------------------------------------------------
