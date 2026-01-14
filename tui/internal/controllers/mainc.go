// Package controllers defines our controller layer
package controllers

import (
	"esdi/peripheral"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"fmt"

	"github.com/rivo/tview"
)

type Ctrls struct {
	MC                  *MainController
	WindowingController *LayoutController
}

type MainController struct {
	App      *tview.Application
	Dom      *dom.DOM
	EvBus    *events.Bus
	DevClerk *peripheral.PeripheralDeviceClerk
}

func NewMainController() *MainController {
	mc := &MainController{
		App:      tview.NewApplication(),
		Dom:      dom.NewDOM(),
		EvBus:    events.NewBus(),
		DevClerk: peripheral.NewPeripheralDeviceClerk(),
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

	mc.EvBus.On(ui.FindCDashDisplay{}, func(e any) {
		err := mc.DevClerk.FindDevices()
		if err != nil {
			mc.EvBus.Emit(ui.LogEv{Log: "Error finding devices: " + err.Error() + "\n"})
		}

		if len(mc.DevClerk.Devices) == 0 {
			mc.EvBus.Emit(ui.LogEv{Log: "  there are no devices\n"})
		}

		for _, d := range mc.DevClerk.Devices {
			msg := fmt.Sprintf("  [%2d] %s\n", d.ID, d.Name)
			mc.EvBus.Emit(ui.LogEv{Log: msg})
		}
	})

	return mc
}
