package ui

import (
	"esdi/cdashdisplay"
	helper "esdi/helpers"
	"esdi/tui/internal/models"

	"github.com/rivo/tview"
)

//

type TUILoaded struct{}

//

// Main controller events

// Find a way to remove this
type ForceRedraw struct{}

type LogEv struct {
	Log string
}

type PrintLogEv struct {
	Log string
}

type RedrawEv struct {
	Fn func()
}

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
	Window models.WindowForm
}

type CreateWindowEv struct {
	Window models.WindowForm
}

type DestroyWindowEv struct {
	ID int16
}

type LayoutRegisterWindowEv struct {
	Window models.WindowForm
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

// + Stream events ------------------------------------------------------------

type StartStreamingReqEv struct{}
type StopStreamingReqEv struct{}

type StreamDataEv struct {
	Str string
}

// - Stream events ------------------------------------------------------------
