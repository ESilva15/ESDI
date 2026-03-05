package controllers

import (
	"esdi/tui/internal/services"
	"esdi/tui/internal/views"

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

	go func() {
		for msg := range stream {
			sc.App.QueueUpdateDraw(func() {
				sc.StreamView.Update(msg)
			})
		}
	}()
}
