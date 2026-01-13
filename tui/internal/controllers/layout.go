// Package controllers
package controllers

import (
	"fmt"

	"esdi/tui/internal/events"
	"esdi/tui/internal/models"
	"esdi/tui/internal/ui"
)

type LayoutController struct {
	EvBus *events.Bus
}

func NewLayoutController(bus *events.Bus) *LayoutController {
	lc := &LayoutController{
		EvBus: bus,
	}

	bus.On(ui.CreateWindowEv{}, func(e any) {
		ev := e.(ui.CreateWindowEv)
		lc.createWindow(ev.Window)
	})

	return lc
}

func (lc *LayoutController) createWindow(win models.Window) {
	lc.EvBus.Emit(ui.LogEv{
		Log: fmt.Sprintf("x     : %d\n"+
			"y     : %d\n"+
			"width : %d\n"+
			"height: %d\n"+
			"title : %s\n", win.X, win.Y, win.Width, win.Height, win.Title),
	})

	lc.EvBus.Emit(ui.WindowCreatedEv{Window: win})
}
