// Package controllers
package controllers

import (
	"esdi/tui/internal/events"
)

type LayoutController struct {
	EvBus *events.Bus
}

func NewLayoutController(bus *events.Bus) *LayoutController {
	lc := &LayoutController{
		EvBus: bus,
	}

	return lc
}

// func (lc *LayoutController) createWindow(win models.Window) {
// 	lc.EvBus.Emit(ui.LogEv{
// 		Log: fmt.Sprintf("x     : %d\n"+
// 			"y     : %d\n"+
// 			"width : %d\n"+
// 			"height: %d\n"+
// 			"title : %s\n", win.X, win.Y, win.Width, win.Height, win.Title),
// 	})
//
// 	lc.EvBus.Emit(ui.WindowCreatedEv{})
// }
