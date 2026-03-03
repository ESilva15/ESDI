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
}

func NewStreamingCtrl(base *Controller, ser *services.CDashService) *StreamingCtrl {
	ctrl := &StreamingCtrl{
		Controller: base,
		Service:    ser,
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
			sc.Stop()
		}

		return ev
	})
}

func (sc *StreamingCtrl) Stop() {
	sc.Service.StopStream()
}

func (sc *StreamingCtrl) Start() {
	sc.Service.StartStream()
	sc.Messages <- "started stream\n"

	stream := sc.Service.GetStream()
	sc.Messages <- "got stream\n"

	go func() {
		for msg := range stream {
			sc.Messages <- "received message\n"
			sc.App.QueueUpdateDraw(func() {
				sc.StreamView.Update(msg)
			})
		}
	}()
}
