// Package controllers defines our controller layer
package controllers

import (
	"esdi/cdashdisplay"
	helper "esdi/helpers"
	"esdi/peripheral"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"esdi/tui/internal/views"
	"fmt"
	"io"

	"github.com/gdamore/tcell/v2"
)

type DeviceController struct {
	*Controller
	DeviceAPIView *views.DeviceAPIView
	DevClerk      *peripheral.PeripheralDeviceClerk
	CDash         *cdashdisplay.CDashDisplay
	StreamStrl    *StreamingCtrl
}

func NewDeviceController(base *Controller) *DeviceController {
	mc := &DeviceController{
		Controller: base,
		DevClerk:   peripheral.NewPeripheralDeviceClerk(),
		CDash:      nil,
		StreamStrl: NewStreamingCtrl(),
	}

	mc.Bus.On(ui.RedrawEv{}, func(e any) {
		go func() {
			rev := e.(ui.RedrawEv)
			if rev.Fn != nil {
				mc.App.QueueUpdateDraw(e.(ui.RedrawEv).Fn)
			}
		}()
	})

	mc.Bus.On(ui.TUILoaded{}, func(e any) {
		mc.Logger.Debug("Yo, we do get here!\n")
		mc.Main()
		mc.PrintToOutputWindow("TUI LOADED EVENT")
	})

	mc.Bus.On(ui.ChangeFocusEv{}, func(e any) {
		mc.App.SetFocus(e.(ui.ChangeFocusEv).Target)
	})

	mc.Bus.On(ui.LogEv{}, func(e any) {
		go func() {
			mc.App.QueueUpdateDraw(func() {
				mc.Bus.Emit(ui.PrintLogEv{Log: e.(ui.LogEv).Log})
			})
		}()
	})

	mc.Bus.On(ui.CreateWindowEv{}, func(e any) {
		win := e.(ui.CreateWindowEv).Window

		winDecor := cdashdisplay.DefaultDecorations
		winDecor.TextSize = win.TextSize
		winDecor.TitleSize = win.TitleSize

		uiWindow := cdashdisplay.UIWindow{
			Dims: cdashdisplay.UIDimensions{
				X0:     win.X,
				Y0:     win.Y,
				Width:  win.Width,
				Height: win.Height,
			},
			Opts: cdashdisplay.UIWindowOpts{
				WinType:      win.Type, // NOTE: values aren't implemented yet
				ShowID:       win.ShowID,
				PreviewValue: helper.B32(win.PreviewValue),
			},
			Decor: winDecor,
			Title: helper.B32(win.Title),
		}

		wID, err := mc.CDash.CreateWindow(uiWindow)
		if err != nil {
			mc.Bus.Emit(ui.PrintLogEv{Log: "failed to create window\n"})
			return
		}

		mc.Bus.Emit(ui.WindowCreatedEv{ID: wID, Win: uiWindow})
		mc.Bus.Emit(ui.PrintLogEv{Log: "Window created!\n"})
	})

	mc.Bus.On(ui.UpdateWindowEv{}, func(e any) {
		winModel := e.(ui.UpdateWindowEv)

		// Build a new UIWindow here I guess
		curWindow, ok := mc.CDash.State.Layout.Windows[winModel.ID]
		if !ok {
			mc.Bus.Emit(ui.PrintLogEv{Log: "could not acquire window from display state"})
			return
		}

		// Need to use a better interface for this me things
		// Update the dimensions
		curWindow.Dims.X0 = winModel.Window.X
		curWindow.Dims.Y0 = winModel.Window.Y
		curWindow.Dims.Width = winModel.Window.Width
		curWindow.Dims.Height = winModel.Window.Height
		// Update the opts
		curWindow.Opts.ShowID = winModel.Window.ShowID
		curWindow.Opts.WinType = winModel.Window.Type
		curWindow.Opts.PreviewValue = helper.B32(winModel.Window.PreviewValue)
		// Update the decorations
		curWindow.Decor.TitleSize = winModel.Window.TitleSize
		curWindow.Decor.TextSize = winModel.Window.TextSize
		// Update the title
		curWindow.Title = helper.B32(winModel.Window.Title)

		err := mc.CDash.UpdateWindow(winModel.ID, curWindow)
		if err != nil {
			mc.Bus.Emit(ui.PrintLogEv{Log: "failed to update window"})
		}
	})

	mc.Bus.On(ui.DestroyWindowEv{}, func(e any) {
		mc.Logger.Info(fmt.Sprintf("Called in to destroy win: %d", e.(ui.DestroyWindowEv).ID))
		mc.CDash.DestroyWindow(e.(ui.DestroyWindowEv).ID)

		mc.Bus.Emit(ui.WindowDestroyedEv{ID: e.(ui.DestroyWindowEv).ID})
	})

	mc.Bus.On(ui.MoveWindowEv{}, func(e any) {
		mvData := e.(ui.MoveWindowEv)
		mc.Logger.Debug(fmt.Sprintf("request to move window '%d'", mvData.WindowID))

		newDims, err := mc.CDash.MoveWindow(mvData.WindowID, mvData.Delta)
		if err != nil && err != io.EOF {
			mc.Logger.Debug(fmt.Sprintf("failed to move window '%d' %s", mvData.WindowID, err.Error()))
			return
		}

		mc.Bus.Emit(ui.WindowMovedEv{ID: mvData.WindowID, Dims: newDims})
	})

	mc.Bus.On(ui.ResizeWindowEv{}, func(e any) {
		mvData := e.(ui.ResizeWindowEv)
		mc.Logger.Debug(fmt.Sprintf("requested to resize window '%d'", mvData.WindowID))
		err := mc.CDash.ResizeWindow(mvData.WindowID, mvData.Delta)
		if err != nil && err != io.EOF {
			mc.Logger.Debug(fmt.Sprintf("failed to resize window '%d' %s", mvData.WindowID, err.Error()))
			return
		}
	})

	mc.Bus.On(ui.SaveLayoutEv{}, func(e any) {
		mc.CDash.SaveLayout()
	})

	mc.Bus.On(ui.LoadLayoutEv{}, func(e any) {
		mc.CDash.LoadLayout()
		mc.Bus.Emit(ui.RegisterLoadedLayout{*mc.CDash.State.Layout})
	})

	mc.Bus.On(ui.ForceRedraw{}, func(e any) {
		// mc.App.QueueUpdateDraw(func() {})
		mc.App.Draw()
	})

	mc.Bus.On(ui.StartStreamingReqEv{}, func(e any) {
		mc.StreamStrl.Start(mc.Bus)
	})

	mc.Bus.On(ui.StopStreamingReqEv{}, func(e any) {
		mc.StreamStrl.Stop(mc.Bus)
	})

	return mc
}

func (mc *DeviceController) PrintToOutputWindow(msg string) {
	mc.App.QueueUpdateDraw(func() {
		fmt.Fprintf(mc.DeviceAPIView.OutputWindow.TextArea,
			"%s", msg)
	})
}

func (mc *DeviceController) findCDashDisplay() {
	mc.PrintToOutputWindow("looking for cdash display\n")
	mc.Logger.Info("Looking for CDashDisplay")

	cdashdisplay.SetLogger(mc.Logger.With("[device]", "cdashdisplay"))

	display, err := cdashdisplay.NewCDashDisplay()
	if err != nil {
		mc.Logger.Info("didn't find cdashdisplay")
		mc.PrintToOutputWindow("didn't find cdash display\n")
		return
	}

	mc.CDash = display
	mc.Logger.Info("found cdashdisplay on: " + display.WT.Cfg.Name)
	mc.PrintToOutputWindow("found cdashdisplay on: " + display.WT.Cfg.Name + "\n")
}

func (mc *DeviceController) setAppEventCapture() {
	mc.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
			mc.App.Stop()
			return nil
		}

		return event
	})
}

func (mc *DeviceController) setDeviceAPIViewEvents() {
	mc.DeviceAPIView.MainFlex.
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Rune() {
			case 'r':
				go mc.findCDashDisplay()
			}
			return ev
		})

}

func (mc *DeviceController) AddDeviceAPIListItems() {
	mc.DeviceAPIView.DevAPIList.AddItem("layout", "build a layout for CDashDisplay",
		func() {
			// var err error
			//
			// // Get the api pages
			// apiToolPages := mc.DeviceAPIView.DevAPIToolView
			//
			// layoutToolUINode := mc.Dom.GetNodeByID(views.LayoutToolFlexID)
			// if layoutToolUINode == nil {
			// 	layoutToolUINode, err = buildLayoutFlexComponent(mc.Bus, mc.Dom)
			// 	if err != nil {
			// 		mc.PrintToOutputWindow(
			// 			fmt.Sprintf("      Failed to build layout tool UI: %s\n", err.Error()),
			// 		)
			// 	}
			// }
			//
			// err = views.AddAndShowPage(apiToolPages.Pages, layoutToolUINode, true)
			// if err == nil {
			// 	mc.App.SetFocus(layoutToolUINode.Self)
			// }
		})
}

func layoutToolUIOnSelect(bus *events.Bus, doc *dom.DOM) {
}

func (mc *DeviceController) mainUI() error {
	var err error

	// Set the main view
	mc.DeviceAPIView, err = views.NewDeviceAPIView(mc.Dom)
	if err != nil {
		return err
	}

	mc.setDeviceAPIViewEvents()

	// Add the API things to the deviceAPIListView
	mc.AddDeviceAPIListItems()

	rootNode, err := mc.Dom.NewUINode("root", nil, mc.DeviceAPIView.MainFlex)
	if err != nil {
		return err
	}

	mc.Dom.SetRoot(rootNode)

	return nil
}

func (mc *DeviceController) Main() error {
	mc.Logger.Debug("and then here")
	err := mc.mainUI()
	mc.Logger.Debug("and then here x2")
	mc.setAppEventCapture()

	// First focused element
	mc.Logger.Debug("and then here x3")
	firstFocus := mc.Dom.GetElemByID(views.DeviceAPIListID)
	if firstFocus == nil {
		return err
	}

	mc.Logger.Debug("and then here x4")
	mc.App.SetRoot(mc.DeviceAPIView.MainFlex, true)
	mc.App.SetFocus(firstFocus)

	return err
}
