package controllers

import (
	"esdi/telemetry"
	"esdi/tui/internal/services"
	"esdi/tui/internal/views"
	"fmt"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
)

type StreamingCtrl struct {
	*Controller
	Service    *services.CDashService
	StreamView *views.StreamView
	Messages   chan string
	Internal   chan string
	Run        bool
	OnExit     func()
	isRunning  bool
	TelemServ  *services.TelemetryService
}

func NewStreamingCtrl(
	base *Controller,
	serCDash *services.CDashService,
	serTelem *services.TelemetryService,
) *StreamingCtrl {
	ctrl := &StreamingCtrl{
		Controller: base,
		Service:    serCDash,
		TelemServ:  serTelem,
		Messages:   make(chan string, 10),
		Internal:   make(chan string, 10),
		Run:        false,
		StreamView: views.NewStreamView(),
		isRunning:  false,
	}

	ctrl.registerHooks()

	return ctrl
}

func (sc *StreamingCtrl) registerHooks() {
	sc.StreamView.TextView.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			sc.OnExit()
		}

		switch ev.Rune() {
		case 's':
			// Start
			sc.Start()
		case 'p':
			// Pause
			// sc.Stop()
		}

		return ev
	})
}

func (sc *StreamingCtrl) Start() {
	stream := sc.TelemServ.StartStream()

	var isDrawing atomic.Bool

	go func() {
		for msg := range stream {
			if isDrawing.Load() {
				continue
			}

			isDrawing.Store(true)

			sc.App.QueueUpdateDraw(func() {
				sc.Messages <- fmt.Sprintf("got data: %+v\n", &msg)
				sc.StreamView.Update(&msg)
				isDrawing.Store(false)
			})
		}
	}()
}

// SetInternalState is used to update the stuff in here, for example, the user
// goes into the layout tool, sets up the data to transmit to his devices and
// then comes here to stream that data. We call this to set the fields the user
// has subscrived to in his tooling
//
// Performance reasoning: this is not used during the high frequency data transmission
// so we can get away with using a map for convenience here
func (sc *StreamingCtrl) SetInternalState() {
	fields := make(map[int16]telemetry.FieldID, len(sc.Service.CDash.State.Layout.Windows))

	for _, w := range sc.Service.CDash.State.Layout.Windows {
		fieldID, _ := telemetry.GetFieldID(w.UIData.TelemetryField)
		fields[w.UIData.IDX] = fieldID
	}

	sc.TelemServ.ActiveProvider.Subscribe(fields)

	sc.Messages <- fmt.Sprintf("Subscribed Fields: %+v [%d]\n", fields, len(fields))
}
