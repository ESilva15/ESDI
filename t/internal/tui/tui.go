// Package tui
package tui

import (
	"esdi/t/internal/controllers"
	dom "esdi/t/internal/dom"
	"esdi/t/internal/events"
	"esdi/t/internal/ui"
	"esdi/t/internal/views"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	App    *tview.Application
	Dom    *dom.DOM
	Events *events.Bus
}

func NewTUI() *TUI {
	return &TUI{
		App:    tview.NewApplication(),
		Dom:    dom.NewDOM(),
		Events: events.NewBus(),
	}
}

func (t *TUI) log(formatter string, args ...any) {
	outputWin := t.Dom.GetElemByID("output-window").(*tview.TextView)
	fmt.Fprintf(outputWin, formatter, args...)
}

func (t *TUI) Start() error {
	// The app starts running here!
	//
	// Set the event capture for the global app itself here
	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
			t.App.Stop()
			return nil
		}

		return event
	})

	ctx := &ui.UIContext{
		Redraw: func() {
			t.App.QueueUpdateDraw(func() {})
		},
		Log: func(formatter string, args ...any) {
			t.log(formatter, args...)
		},
		ChangeFocus: func(p tview.Primitive) {
			t.App.SetFocus(p)
		},
	}

	// Create the controllers
	windowingController := &controllers.WindowingController{
		Events: t.Events,
		Ctx:    ctx,
	}

	views.BindWindowEvents(ctx, t.Events, nil)

	// Make the output window
	err := views.BuildOutputWindow(t.Dom, ctx)
	if err != nil {
		panic("failed to create output window")
	}

	// Set the main view
	mainUINode, err := views.BuildMainFlex(t.Dom, ctx, windowingController)
	if err != nil {
		panic(fmt.Errorf("failed to build main flex - %s", err.Error()))
	}

	rootNode, err := t.Dom.NewUINode("root", nil, mainUINode)
	if err != nil {
		panic(fmt.Errorf("failed to create root node - %s", err.Error()))
	}

	t.Dom.SetRoot(rootNode)
	firstFocus := t.Dom.GetElemByID("device-api-list")
	if firstFocus == nil {
		panic(fmt.Errorf("`list-window` isn't registered"))
	}

	// Star the app
	err = t.App.SetRoot(t.Dom.GetRootElem(), true).SetFocus(firstFocus).Run()
	if err != nil {
		return err
	}

	return nil
}
