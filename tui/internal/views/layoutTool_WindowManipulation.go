package views

import (
	"fmt"

	helper "esdi/helpers"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ManipMode int

const (
	moveMode ManipMode = iota
	resizeMode
)

type windowManipState struct {
	Mode ManipMode
}

func keyToVector(r rune) (helper.Vector, bool) {
	var mult uint16 = 1

	// Is uppercase? For fast movement
	if r >= 'A' && r <= 'Z' {
		mult = 10
		r += 'a' - 'A'
	}

	switch r {
	case 'h':
		return helper.Vector{DX: -mult, DY: 0}, true
	case 'j':
		return helper.Vector{DX: 0, DY: mult}, true
	case 'k':
		return helper.Vector{DX: 0, DY: -mult}, true
	case 'l':
		return helper.Vector{DX: mult, DY: 0}, true
	}

	return helper.Vector{}, false
}

func handleMovementCapture(bus *events.Bus, id int16, ev *tcell.EventKey) *tcell.EventKey {
	vec, ok := keyToVector(ev.Rune())
	if !ok {
		return ev
	}

	bus.Emit(ui.MoveWindowEv{
		WindowID: id,
		Delta:    vec,
	})

	return ev
}

func handleResizeCapture(bus *events.Bus, id int16, ev *tcell.EventKey) *tcell.EventKey {
	vec, ok := keyToVector(ev.Rune())
	if !ok {
		return ev
	}

	bus.Emit(ui.ResizeWindowEv{
		WindowID: id,
		Delta:    vec,
	})

	return ev
}

type modeHandler func(*tcell.EventKey) *tcell.EventKey

func (s *windowManipState) CurrentHandler(
	bus *events.Bus,
	id int16,
) modeHandler {
	switch s.Mode {
	case moveMode:
		return func(ev *tcell.EventKey) *tcell.EventKey {
			return handleMovementCapture(bus, id, ev)
		}
	case resizeMode:
		return func(ev *tcell.EventKey) *tcell.EventKey {
			return handleResizeCapture(bus, id, ev)
		}
	}

	return nil
}

func windowManipulationEvCapture(
	bus *events.Bus,
	doc *dom.DOM,
	id int16,
	state *windowManipState,
) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEscape:
			bus.Emit(ui.ChangeFocusEv{
				Target: doc.GetElemByID(layoutToolFlexID),
			})
			return ev
		}

		bus.Emit(ui.PrintLogEv{
			Log: fmt.Sprintf("-> IDX: %d\n", id),
		})

		// Mode switching
		switch ev.Rune() {
		case 'r':
			state.Mode = resizeMode
			bus.Emit(ui.PrintLogEv{Log: "switched to resizeMode"})
		case 'm':
			state.Mode = moveMode
			bus.Emit(ui.PrintLogEv{Log: "switched to moveMode"})
		}

		// Delegate the current handler input
		handler := state.CurrentHandler(bus, id)
		if handler != nil {
			return handler(ev)
		}

		return ev
	}
}

func windowManipulationTool(bus *events.Bus, doc *dom.DOM, idx int16) {
	bus.Emit(ui.PrintLogEv{Log: fmt.Sprintf("WID: %d\n", idx)})

	// Get the ActionPages parent
	box := tview.NewBox().SetTitle("MOVE WINDOW TOOL")
	box.SetBorder(true)

	var err error
	var boxNode *dom.UINode

	// delete the currently existing move-window-box
	actionPages := doc.GetElemByID(layoutToolActionPagesID).(*tview.Pages)
	actionPages.RemovePage("move-window-box")

	boxNode = doc.GetNodeByID("move-window-box")
	if boxNode != nil {
		doc.DeleteElem(boxNode)
	}

	boxNode, err = doc.NewUINode(
		"move-window-box",
		doc.GetElemByID(layoutToolActionPagesID),
		box,
	)

	if err != nil {
		return
	}

	// Create the state for this tool
	state := windowManipState{
		Mode: moveMode,
	}

	box.SetInputCapture(windowManipulationEvCapture(bus, doc, idx, &state))
	AddAndShowPage(bus, doc, actionPages, boxNode, true)
}
