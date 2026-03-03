// Package controllers defines our controller layer
package controllers

import (
	serv "esdi/tui/internal/services"
	"esdi/tui/internal/views"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type DeviceController struct {
	*Controller
	DeviceAPIView *views.DeviceAPIView
	LayoutCtrl    *LayoutController
	StreamCtrl    *StreamingCtrl
	DevService    *serv.CDashService
}

func NewDeviceController(base *Controller, devService *serv.CDashService) *DeviceController {
	mc := &DeviceController{
		Controller: base,
		LayoutCtrl: NewLayoutController(base, devService),
		DevService: devService,
		StreamCtrl: NewStreamingCtrl(base, devService),
	}

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
		fmt.Fprintf(mc.DeviceAPIView.OutputWindow.TextArea, "%s", msg)
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
	mc.DeviceAPIView.DevAPIList.
		AddItem("layout", "build a layout for CDashDisplay", func() {
			// Get the api pages
			views.AddAndShowPage(
				mc.DeviceAPIView.DevAPIToolView.Pages,
				"layout-tool",
				mc.LayoutCtrl.LayoutToolView.Flex,
			)
			mc.App.SetFocus(mc.LayoutCtrl.LayoutToolView.LayoutTree.Tree)
		})
	mc.DeviceAPIView.DevAPIList.
		AddItem("stream", "stream data to the display", func() {
			views.AddAndShowPage(mc.DeviceAPIView.DevAPIToolView.Pages,
				"streaming-tool",
				mc.StreamCtrl.StreamView.TextView,
			)
			mc.App.SetFocus(mc.StreamCtrl.StreamView.TextView)
		})
}

func (mc *DeviceController) injectViewCallbacks() {
	mc.setDeviceAPIViewEvents()
}

func (mc *DeviceController) injectControllerCallbacks() {
	// Tell the layout controller what to do on exit
	giveFocusToAPIList := func() {
		mc.App.SetFocus(mc.DeviceAPIView.DevAPIList.List)
	}

	mc.LayoutCtrl.OnExit = giveFocusToAPIList
	mc.StreamCtrl.OnExit = giveFocusToAPIList
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

	go func() {
		for msg := range mc.StreamCtrl.Messages {
			mc.PrintToOutputWindow(msg)
		}
	}()
}

func (mc *DeviceController) mainUI() error {
	var err error

	// Set the main view
	mc.DeviceAPIView, err = views.NewDeviceAPIView()
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
