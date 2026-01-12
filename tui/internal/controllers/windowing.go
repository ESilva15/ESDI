// Package controllers
package controllers

import (
	"esdi/tui/internal/events"
	"esdi/tui/internal/models"
	"esdi/tui/internal/ui"
	"strconv"
)

type WindowingController struct {
	Ctx    *ui.UIContext
	Events *events.Bus
}

func (wc *WindowingController) CreateWindow(x, y,
	width, height, title string) {
	wc.Ctx.Log("x     : %s\n", x)
	wc.Ctx.Log("y     : %s\n", y)
	wc.Ctx.Log("width : %s\n", width)
	wc.Ctx.Log("height: %s\n", height)
	wc.Ctx.Log("title : %s\n", title)
	// newWindow := tview.NewTreeNode(title)

	xValue, err := strconv.ParseUint(x, 10, 64)
	if err != nil {
		wc.Events.Emit(events.Error{Error: err})
		return
	}
	yValue, err := strconv.ParseUint(y, 10, 64)
	if err != nil {
		wc.Events.Emit(events.Error{Error: err})
		return
	}
	widthValue, err := strconv.ParseUint(width, 10, 64)
	if err != nil {
		wc.Events.Emit(events.Error{Error: err})
		return
	}
	heightValue, err := strconv.ParseUint(height, 10, 64)
	if err != nil {
		wc.Events.Emit(events.Error{Error: err})
		return
	}

	window := models.Window{
		X:      uint16(xValue),
		Y:      uint16(yValue),
		Width:  uint16(widthValue),
		Height: uint16(heightValue),
		Title:  title,
	}

	wc.Events.Emit(events.WindowCreated{Window: window})
}
