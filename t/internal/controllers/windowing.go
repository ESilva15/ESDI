// Package controllers
package controllers

import (
	"esdi/t/internal/events"
	"esdi/t/internal/models"
	"esdi/t/internal/ui"
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

	xValue, err := strconv.ParseInt(x, 10, 1)
	if err != nil {
		// return err
	}
	yValue, err := strconv.ParseInt(y, 10, 1)
	if err != nil {
		// return err
	}
	widthValue, err := strconv.ParseInt(width, 10, 1)
	if err != nil {
		// return err
	}
	heightValue, err := strconv.ParseInt(height, 10, 1)
	if err != nil {
		// return err
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
