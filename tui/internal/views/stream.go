package views

import (
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Car data lengths
const (
	SpeedLen     = 5
	GearLen      = 3
	RpmLen       = 6
	BrakeBiasLen = 6
)

const (
	streamingBoxID = "streaming-box"
)

func streamingWindow(bus *events.Bus, doc *dom.DOM) {
	bus.Emit(ui.LogEv{Log: "creating streaming window"})

	// Get the ActionPages parent
	box := tview.NewTextView()
	box.SetBorder(true).SetTitle("STREAMING")
	box.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			// well, I really need to design the UI (as if lmao) better
			os.Exit(0)
		}

		return ev
	})

	var err error
	var boxNode *dom.UINode

	// delete the currently existing streaming-box
	actionPages := doc.GetElemByID(LayoutToolActionPagesID).(*tview.Pages)
	actionPages.RemovePage(streamingBoxID)

	// Get the currently exisiting streaming box in the dom so we can delete it
	boxNode = doc.GetNodeByID(streamingBoxID)
	if boxNode != nil {
		doc.DeleteElem(boxNode)
	}

	// Register this streaming box as an UINode
	boxNode, err = doc.NewUINode(
		streamingBoxID,
		doc.GetElemByID(LayoutToolActionPagesID),
		box,
	)
	if err != nil {
		return
	}

	// Add it to the action pages and view it
	AddAndShowPage(actionPages, boxNode, true)

	bus.Emit(ui.StartStreamingReqEv{})

	bus.On(ui.StreamDataEv{}, func(e any) {
		bus.Emit(ui.RedrawEv{
			Fn: func() {
				sd := e.(ui.StreamDataEv)
				box.SetText(sd.Str)
			},
		})
	})
}
