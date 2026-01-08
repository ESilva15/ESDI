// Package tui
package tui

import (
	dom "esdi/t/internal/dom"
	"esdi/t/internal/views"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	App *tview.Application
	Dom *dom.DOM
}

var (
	application *TUI
	initiated   = false
)

func NewTUI() *TUI {
	return &TUI{
		App: tview.NewApplication(),
		Dom: dom.NewDOM(),
	}
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

	ctx := &views.UIContext{
		Redraw: func() {
			t.App.QueueUpdateDraw(func() {})
		},
	}

	// Set the main view
	mainUINode, err := views.BuildMainFlex(t.Dom, ctx)
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

func Init() {
	initiated = true
}
