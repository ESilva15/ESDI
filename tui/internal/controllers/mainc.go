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
	"io"
	"log/slog"

	"github.com/rivo/tview"
)

var pLogger *slog.Logger

func mcLog(msg string, args ...any) {
	pLogger.Debug(msg, args)
}

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
	pLogger = logger
	mc := &MainController{
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

			mc.EvBus.Emit(ui.WindowCreatedEv{ID: wID, Win: uiWindow})
			mc.EvBus.Emit(ui.PrintLogEv{Log: "Window created!\n"})
		}()
	})

	mc.EvBus.On(ui.UpdateWindowEv{}, func(e any) {
		winModel := e.(ui.UpdateWindowEv)

		// Build a new UIWindow here I guess
		curWindow, ok := mc.CDash.State.Layout.Windows[winModel.ID]
		if !ok {
			mc.EvBus.Emit(ui.PrintLogEv{Log: "could not acquire window from display state"})
			return
		}

		curWindow.Dims.X0 = winModel.Window.X
		curWindow.Dims.Y0 = winModel.Window.Y
		curWindow.Dims.Width = winModel.Window.Width
		curWindow.Dims.Height = winModel.Window.Height
		curWindow.Title = helper.B32(winModel.Window.Title)

		err := mc.CDash.UpdateWindow(winModel.ID, curWindow)
		if err != nil {
			mc.EvBus.Emit(ui.PrintLogEv{Log: "failed to update window"})
		}
	})

	mc.EvBus.On(ui.DestroyWindowEv{}, func(e any) {
		go func() {
			pLogger.Info(fmt.Sprintf("Called in to destroy win: %d", e.(ui.DestroyWindowEv).ID))
			mc.CDash.DestroyWindow(e.(ui.DestroyWindowEv).ID)

			mc.EvBus.Emit(ui.WindowDestroyedEv{ID: e.(ui.DestroyWindowEv).ID})
		}()
	})

	mc.EvBus.On(ui.MoveWindowEv{}, func(e any) {
		mvData := e.(ui.MoveWindowEv)
		pLogger.Debug(fmt.Sprintf("request to move window '%d'", mvData.WindowID))
		err := mc.CDash.MoveWindow(mvData.WindowID, mvData.Delta)
		if err != nil && err != io.EOF {
			pLogger.Debug(fmt.Sprintf("failed to move window '%d' %s", mvData.WindowID, err.Error()))
			return
		}
	})

	mc.EvBus.On(ui.ResizeWindowEv{}, func(e any) {
		mvData := e.(ui.ResizeWindowEv)
		pLogger.Debug(fmt.Sprintf("requested to resize window '%d'", mvData.WindowID))
		err := mc.CDash.ResizeWindow(mvData.WindowID, mvData.Delta)
		if err != nil && err != io.EOF {
			pLogger.Debug(fmt.Sprintf("failed to resize window '%d' %s", mvData.WindowID, err.Error()))
			return
		}
	})

	mc.EvBus.On(ui.SaveLayoutEv{}, func(e any) {
		mc.CDash.SaveLayout()
	})

	mc.EvBus.On(ui.LoadLayoutEv{}, func(e any) {
		mc.CDash.LoadLayout()
		mc.EvBus.Emit(ui.RegisterLoadedLayout{*mc.CDash.State.Layout})
	})

	mc.EvBus.On(ui.FindCDashDisplay{}, func(e any) {
		go func() {
			mc.EvBus.Emit(ui.PrintLogEv{Log: "looking for cdash display\n"})
			pLogger.Info("Looking for CDashDisplay")

			cdashdisplay.SetLogger(pLogger.With("[device]", "cdashdisplay"))

			display, err := cdashdisplay.NewCDashDisplay()
			if err != nil {
				pLogger.Info("didn't find cdashdisplay")
				mc.EvBus.Emit(ui.PrintLogEv{Log: "didn't find cdash display\n"})
				return
			}

			// pLogger.Info(fmt.Sprintf("found cdashdisplay on %s!", mc.CDash.WT.Cfg.Name))
			// mc.EvBus.Emit(ui.PrintLogEv{Log: "found cdashdisplay" + mc.CDash.WT.Cfg.Name + "\n"})
			mc.CDash = display
		}()
	})

	mc.EvBus.On(ui.ForceRedraw{}, func(e any) {
		mc.App.QueueUpdateDraw(func() {})
	})

	return mc
}
