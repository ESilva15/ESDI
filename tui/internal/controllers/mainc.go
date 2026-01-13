// Package controllers defines our controller layer
package controllers

import (
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"

	"github.com/rivo/tview"
)

type Ctrls struct {
	MC                  *MainController
	WindowingController *LayoutController
}

type MainController struct {
	App   *tview.Application
	Dom   *dom.DOM
	EvBus *events.Bus
}

func NewMainController() *MainController {
	mc := &MainController{
		App:   tview.NewApplication(),
		Dom:   dom.NewDOM(),
		EvBus: events.NewBus(),
	}

	mc.EvBus.On(ui.RedrawEv{}, func(e any) {
		mc.App.QueueUpdateDraw(func() {})
	})

	mc.EvBus.On(ui.ChangeFocusEv{}, func(e any) {
		mc.App.SetFocus(e.(ui.ChangeFocusEv).Target)
	})

	mc.EvBus.On(ui.LogEv{}, func(e any) {
		mc.EvBus.Emit(ui.PrintLogEv{Log: e.(ui.LogEv).Log})
	})

	return mc
}
