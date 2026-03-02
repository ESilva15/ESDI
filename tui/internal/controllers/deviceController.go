// Package controllers defines our controller layer
package controllers

import (
	serv "esdi/tui/internal/services"
	"esdi/tui/internal/ui"
	"esdi/tui/internal/views"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type DeviceController struct {
	*Controller
	DeviceAPIView *views.DeviceAPIView
	LayoutCtrl    *LayoutController
	StreamStrl    *StreamingCtrl
	DevService    *serv.CDashService
}

func NewDeviceController(base *Controller, devService *serv.CDashService) *DeviceController {
	mc := &DeviceController{
		Controller: base,
		LayoutCtrl: NewLayoutController(base, devService),
		DevService: devService,
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

	// mc.Bus.On(ui.CreateWindowEv{}, func(e any) {
	// })

	// mc.Bus.On(ui.UpdateWindowEv{}, func(e any) {
	// 	winModel := e.(ui.UpdateWindowEv)
	//
	// 	// Build a new UIWindow here I guess
	// 	curWindow, ok := mc.CDash.State.Layout.Windows[winModel.ID]
	// 	if !ok {
	// 		mc.Bus.Emit(ui.PrintLogEv{Log: "could not acquire window from display state"})
	// 		return
	// 	}
	//
	// 	// Need to use a better interface for this me things
	// 	// Update the dimensions
	// 	curWindow.Dims.X0 = winModel.Window.X
	// 	curWindow.Dims.Y0 = winModel.Window.Y
	// 	curWindow.Dims.Width = winModel.Window.Width
	// 	curWindow.Dims.Height = winModel.Window.Height
	// 	// Update the opts
	// 	curWindow.Opts.ShowID = winModel.Window.ShowID
	// 	curWindow.Opts.WinType = winModel.Window.Type
	// 	curWindow.Opts.PreviewValue = helper.B32(winModel.Window.PreviewValue)
	// 	// Update the decorations
	// 	curWindow.Decor.TitleSize = winModel.Window.TitleSize
	// 	curWindow.Decor.TextSize = winModel.Window.TextSize
	// 	// Update the title
	// 	curWindow.Title = helper.B32(winModel.Window.Title)
	//
	// 	err := mc.CDash.UpdateWindow(winModel.ID, curWindow)
	// 	if err != nil {
	// 		mc.Bus.Emit(ui.PrintLogEv{Log: "failed to update window"})
	// 	}
	// })

	// mc.Bus.On(ui.DestroyWindowEv{}, func(e any) {
	// 	mc.Logger.Info(fmt.Sprintf("Called in to destroy win: %d", e.(ui.DestroyWindowEv).ID))
	// 	mc.CDash.DestroyWindow(e.(ui.DestroyWindowEv).ID)
	//
	// 	mc.Bus.Emit(ui.WindowDestroyedEv{ID: e.(ui.DestroyWindowEv).ID})
	// })

	// mc.Bus.On(ui.MoveWindowEv{}, func(e any) {
	// 	mvData := e.(ui.MoveWindowEv)
	// 	mc.Logger.Debug(fmt.Sprintf("request to move window '%d'", mvData.WindowID))
	//
	// 	newDims, err := mc.CDash.MoveWindow(mvData.WindowID, mvData.Delta)
	// 	if err != nil && err != io.EOF {
	// 		mc.Logger.Debug(fmt.Sprintf("failed to move window '%d' %s", mvData.WindowID, err.Error()))
	// 		return
	// 	}
	//
	// 	mc.Bus.Emit(ui.WindowMovedEv{ID: mvData.WindowID, Dims: newDims})
	// })

	// mc.Bus.On(ui.ResizeWindowEv{}, func(e any) {
	// 	mvData := e.(ui.ResizeWindowEv)
	// 	mc.Logger.Debug(fmt.Sprintf("requested to resize window '%d'", mvData.WindowID))
	// 	err := mc.CDash.ResizeWindow(mvData.WindowID, mvData.Delta)
	// 	if err != nil && err != io.EOF {
	// 		mc.Logger.Debug(fmt.Sprintf("failed to resize window '%d' %s", mvData.WindowID, err.Error()))
	// 		return
	// 	}
	// })

	// mc.Bus.On(ui.SaveLayoutEv{}, func(e any) {
	// 	mc.CDash.SaveLayout()
	// })

	// mc.Bus.On(ui.ForceRedraw{}, func(e any) {
	// 	// mc.App.QueueUpdateDraw(func() {})
	// 	mc.App.Draw()
	// })

	// mc.Bus.On(ui.StartStreamingReqEv{}, func(e any) {
	// 	mc.StreamStrl.Start(mc.Bus)
	// })

	// mc.Bus.On(ui.StopStreamingReqEv{}, func(e any) {
	// 	mc.StreamStrl.Stop(mc.Bus)
	// })

	return mc
}

func (mc *DeviceController) PrintToOutputWindow(msg string) {
	mc.App.QueueUpdateDraw(func() {
		fmt.Fprintf(mc.DeviceAPIView.OutputWindow.TextArea,
			"%s", msg)
	})
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
	// api list hooks
	mc.DeviceAPIView.DevAPIList.List.
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Rune() {
			case 'r':
				go mc.DevService.FindDevice()
			}
			return ev
		})
}

func (mc *DeviceController) AddDeviceAPIListItems() {
	mc.DeviceAPIView.DevAPIList.AddItem("layout", "build a layout for CDashDisplay",
		func() {
			// Get the api pages
			views.AddAndShowPage(
				mc.DeviceAPIView.DevAPIToolView.Pages,
				"layout-tool",
				mc.LayoutCtrl.LayoutToolView.Flex,
			)
			mc.App.SetFocus(mc.LayoutCtrl.LayoutToolView.LayoutTree.Tree)
		})
}

func (mc *DeviceController) injectViewCallbacks() {
	mc.setDeviceAPIViewEvents()
}

func (mc *DeviceController) injectControllerCallbacks() {
	// Tell the layout controller what to do on exit
	mc.LayoutCtrl.OnExit = func() {
		mc.App.SetFocus(mc.DeviceAPIView.DevAPIList.List)
	}
}

func (mc *DeviceController) injectChannels() {
	go func() {
		for msg := range mc.DevService.Messages {
			mc.PrintToOutputWindow(msg)
		}
	}()

	go func() {
		for msg := range mc.LayoutCtrl.Messages {
			mc.PrintToOutputWindow(msg)
		}
	}()
}

func (mc *DeviceController) mainUI() error {
	var err error

	// Set the main view
	mc.DeviceAPIView, err = views.NewDeviceAPIView(mc.Dom)
	if err != nil {
		return err
	}

	// Inject the view hooks
	mc.injectViewCallbacks()

	// Inject controller callbacks
	mc.injectControllerCallbacks()

	// Run channels
	mc.injectChannels()

	// Add the API things to the deviceAPIListView
	mc.AddDeviceAPIListItems()

	return nil
}

func (mc *DeviceController) Main() error {
	err := mc.mainUI()
	mc.setAppEventCapture()

	mc.App.SetRoot(mc.DeviceAPIView.MainFlex, true)
	mc.App.SetFocus(mc.DeviceAPIView.DevAPIList.List)

	return err
}
