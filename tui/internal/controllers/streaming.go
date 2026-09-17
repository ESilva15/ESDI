package controllers

import (
	"fmt"
	"log/slog"
	"sync/atomic"

	"esdi/config"
	"esdi/devices/uidevice"
	"esdi/providers"
	"esdi/services"
	"esdi/telemetry"
	"esdi/tui/internal/models"
	"esdi/tui/internal/views"

	"github.com/gdamore/tcell/v2"
)

type StreamingCtrl struct {
	*Controller
	Service     *services.DeviceService
	StreamView  *views.StreamToolView
	Messages    chan string
	Internal    chan string
	TelemetryCh <-chan telemetry.TelemetryData
	Run         bool
	OnExit      func()
	TelemServ   *services.TelemetryService

	// Stream State
	isRunning bool
}

func NewStreamingCtrl(
	base *Controller,
	devService *services.DeviceService,
	serTelem *services.TelemetryService,
) *StreamingCtrl {
	// NOTE: looks sus, put this somewhere also. Not very good in here
	providerList := []providers.Provider{}
	for _, item := range providers.Providers {
		providerList = append(providerList, item)
	}
	streamView := views.NewStreamToolView(providerList, config.GetCfg().DefaultSim)

	ctrl := &StreamingCtrl{
		Controller:  base,
		Service:     devService,
		TelemServ:   serTelem,
		Messages:    make(chan string, 10),
		Internal:    make(chan string, 10),
		TelemetryCh: make(chan telemetry.TelemetryData, 1),
		Run:         false,
		StreamView:  streamView,
		isRunning:   false,
	}

	ctrl.registerHooks()
	// ctrl.subscribeListeners()

	return ctrl
}

// func (sc *StreamingCtrl) subscribeListeners() {
// 	// Here I will set a UIDevice
// 	sc.TelemetryCh = sc.TelemServ.SubscribeListener("UI", 1)
// }

func (sc *StreamingCtrl) registerHooks() {
	sc.StreamView.Options.Form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			sc.OnExit()
		}

		switch ev.Rune() {
		case 's':
			// Start - stop
			sc.StartStop()
		case 'u':
			// Update
			sc.updateStream()
		}

		return ev
	})

	// Set the options form callbacks
	err := SetFormButtonCallback(sc.StreamView.Options.Form, "Update", func() {
		sc.updateStream()
	})
	if err != nil {
		// NOTE: what do we do here?
		panic("Failed to set callback for streaming controller form update button")
	}
}

func (sc *StreamingCtrl) StartStop() {
	if sc.isRunning {
		slog.Info("stopping stream")

		sc.TelemServ.StopStream()
		sc.Service.StopStream()

		sc.isRunning = false
		return
	}

	// stream is not running, we have to start it now
	// NOTE:
	// Subscribe the only existing device - needs to be discovered by now
	slog.Debug("setting the data stream for device servie")
	sc.Service.SetTelemetryChannel(sc.TelemServ.SubscribeListener("DeviceService", 1))

	dev, err := sc.Service.GetDevice(uidevice.NAME)
	if err == nil {
		if uiDev, ok := dev.(*uidevice.UIDevice); ok {
			sc.TelemetryCh = uiDev.DataChannel()
			go sc.listenToUIStream()
		}
	}

	slog.Debug("starting services")
	sc.Service.StartStream()
	sc.TelemServ.StartStream()

	sc.isRunning = true

	slog.Debug("starting stream")
}

func (sc *StreamingCtrl) parseStreamUpdateForm(form *views.StreamOptionsView) (*models.StreamOptions, error) {
	_, sim := form.SimDropdown.GetCurrentOption()

	return &models.StreamOptions{
		Sim: sim,
	}, nil
}

// updateStream will get the new options from the form and update the state accordingly
func (sc *StreamingCtrl) updateStream() {
	sc.Messages <- "Request to update stream received - parsing form"

	formData, err := sc.parseStreamUpdateForm(sc.StreamView.Options)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to parse form data: %+v", err))
		return
	}

	// NOTE: find a way to hold state so we can compare this new state to the old
	// state. Need a streamctrl state that holds the form model, for example

	slog.Debug(fmt.Sprintf("Parsed form data: %+v", formData))
}

// SetInternalState is used to update the stuff in here, for example, the user
// goes into the layout tool, sets up the data to transmit to his devices and
// then comes here to stream that data. We call this to set the fields the user
// has subscrived to in his tooling
//
// Performance reasoning: this is not used during the high frequency data transmission
// so we can get away with using a map for convenience here
func (sc *StreamingCtrl) SetInternalState() {
	// Acquire the cdashdisplay
	// displayIF, err := sc.Service.GetDevice(cdashdisplay.NAME)
	// if err != nil {
	// 	sc.Messages <- "failed to get " + cdashdisplay.NAME
	// 	return
	// }
	// display, ok := displayIF.(*cdashdisplay.CDashDisplay)
	// if !ok {
	// 	sc.Messages <- "failed to acquire " + cdashdisplay.NAME
	// 	return
	// }
	// ---

	sc.TelemServ.SubscribeToFields()

	// sc.Messages <- fmt.Sprintf("Subscribed Fields: %+v [%d]\n", fields, len(fields))
	// Should I update this?
	sc.Messages <- fmt.Sprintf("Subscribed to fields\n")
}

func (sc *StreamingCtrl) listenToUIStream() {
	var isDrawing atomic.Bool

	for msg := range sc.TelemetryCh {
		if isDrawing.Load() {
			continue
		}
		isDrawing.Store(true)

		// sc.Logger.Debug("got data", "data", msg)

		// Capture locally
		telemetryMsg := msg

		sc.App.QueueUpdateDraw(func() {
			sc.StreamView.Visualizer.Update(&telemetryMsg)
			isDrawing.Store(false)
		})
	}
}
