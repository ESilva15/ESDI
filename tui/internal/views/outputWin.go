package views

import (
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"fmt"

	"github.com/rivo/tview"
)

func bindOutputWindowEvents(
	bus *events.Bus,
	output *tview.TextView,
) {
	bus.On(ui.PrintLogEv{}, func(e any) {
		le, _ := e.(ui.PrintLogEv)
		fmt.Fprintf(output, le.Log)
	})
}

func BuildOutputWindow(bus *events.Bus, doc *dom.DOM) error {
	var outputWin *tview.TextView
	outputWin = tview.NewTextView().SetChangedFunc(func() {
		bus.Emit(ui.RedrawEv{})
		outputWin.ScrollToEnd()
	})

	outputWin.SetBorder(true).SetTitle("DebugWindow")
	_, err := doc.NewUINode("output-window", nil, outputWin)
	if err != nil {
		return err
	}

	bindOutputWindowEvents(bus, outputWin)

	return nil
}
