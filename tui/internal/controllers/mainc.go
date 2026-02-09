// Package controllers defines our controller layer
package controllers

import (
	"esdi/cdashdisplay"
	helper "esdi/helpers"
	"esdi/peripheral"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"fmt"
	"log/slog"

	"github.com/rivo/tview"
)

type Ctrls struct {
	MC                  *MainController
	WindowingController *LayoutController
}

type MainController struct {
	Logger   *slog.Logger
	App      *tview.Application
	Dom      *dom.DOM
	EvBus    *events.Bus
	DevClerk *peripheral.PeripheralDeviceClerk
	CDash    *cdashdisplay.CDashDisplay
}

func NewMainController(logger *slog.Logger) *MainController {
	mc := &MainController{
		Logger:   logger,
		App:      tview.NewApplication(),
		Dom:      dom.NewDOM(),
		EvBus:    events.NewBus(),
		DevClerk: peripheral.NewPeripheralDeviceClerk(),
		CDash:    nil,
	}

	mc.EvBus.On(ui.RedrawEv{}, func(e any) {
		go func() {
			mc.App.QueueUpdateDraw(func() {})
		}()
	})

	mc.EvBus.On(ui.ChangeFocusEv{}, func(e any) {
		go func() {
			mc.App.SetFocus(e.(ui.ChangeFocusEv).Target)
		}()
	})

	mc.EvBus.On(ui.LogEv{}, func(e any) {
		go func() {
			mc.EvBus.Emit(ui.PrintLogEv{Log: e.(ui.LogEv).Log})
		}()
	})

	mc.EvBus.On(ui.CreateWindowEv{}, func(e any) {
		go func() {
			win := e.(ui.CreateWindowEv).Window

			uiWindow := cdashdisplay.UIWindow{
				Dims: cdashdisplay.UIDimensions{
					X0:     win.X,
					Y0:     win.Y,
					Width:  win.Width,
					Height: win.Height,
				},
				Decor: cdashdisplay.DefaultDecorations,
				Title: helper.B32(win.Title),
			}

			wID, err := mc.CDash.CreateWindow(uiWindow)
			if err != nil {
				mc.EvBus.Emit(ui.PrintLogEv{Log: "failed to create window\n"})
				return
			}

			mc.EvBus.Emit(ui.WindowCreatedEv{ID: wID, Title: win.Title})
			mc.EvBus.Emit(ui.PrintLogEv{Log: "Window created!\n"})
		}()
	})

	mc.EvBus.On(ui.DestroyWindowEv{}, func(e any) {
		mc.Logger.Info(fmt.Sprintf("Called in to destroy win: %d", e.(ui.DestroyWindowEv).ID))
		mc.CDash.DestroyWindow(e.(ui.DestroyWindowEv).ID)
	})

	mc.EvBus.On(ui.FindCDashDisplay{}, func(e any) {
		go func() {
			mc.Logger.Info("Looking for CDashDisplay")
			cdashdisplay.SetLogger(mc.Logger.With("[device]", "cdashdisplay"))

			display, err := cdashdisplay.NewCDashDisplay()
			if err != nil {
				return
			}

			mc.CDash = display
		}()
	})

	return mc
}
