package ui

import (
	"esdi/cdashdisplay"
	helper "esdi/helpers"
	"esdi/tui/internal/models"

	"github.com/rivo/tview"
)

// Main controller events

// Find a way to remove this
type ForceRedraw struct{}

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

type SaveLayoutEv struct{}
type LoadLayoutEv struct{}
type RegisterLoadedLayout struct {
	Layout cdashdisplay.LayoutTree
}

type UpdateWindowEv struct {
	ID     int16
	Window models.Window
}

type CreateWindowEv struct {
	Window models.Window
}

type DestroyWindowEv struct {
	ID int16
}

type LayoutRegisterWindowEv struct {
	Window models.Window
}

type MoveWindowEv struct {
	WindowID int16
	Delta    helper.Vector
}

type ResizeWindowEv struct {
	WindowID int16
	Delta    helper.Vector
}

type WindowCreatedEv struct {
	ID  int16
	Win cdashdisplay.UIWindow
}

type WindowMovedEv struct {
	ID   int16
	Dims cdashdisplay.UIDimensions
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
