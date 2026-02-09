// Package tui
package tui

import (
	"esdi/tui/internal/controllers"
	"esdi/tui/internal/views"
	"fmt"
	"log/slog"

	"github.com/gdamore/tcell/v2"
)

func Start(logger *slog.Logger) error {
	// The app starts running here!
	//
	// Set the event capture for the global app itself here
	mc := controllers.NewMainController(logger.With("[ctrl]", "main"))

	mc.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
			mc.App.Stop()
			return nil
		}

		return event
	})

	controllers.NewLayoutController(mc.EvBus)

	// Set the main view
	mainUINode, err := views.BuildMainFlex(mc.EvBus, mc.Dom)
	if err != nil {
		panic(fmt.Errorf("failed to build main flex - %s", err.Error()))
	}

	rootNode, err := mc.Dom.NewUINode("root", nil, mainUINode)
	if err != nil {
		panic(fmt.Errorf("failed to create root node - %s", err.Error()))
	}

	mc.Dom.SetRoot(rootNode)
	firstFocus := mc.Dom.GetElemByID("device-api-list")
	if firstFocus == nil {
		panic(fmt.Errorf("`list-window` isn't registered"))
	}

	// Star the app
	err = mc.App.SetRoot(mc.Dom.GetRootElem(), true).SetFocus(firstFocus).Run()
	if err != nil {
		return err
	}

	return nil
}
