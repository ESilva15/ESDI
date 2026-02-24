package views

import (
	"github.com/rivo/tview"
)

// func bindOutputWindowEvents(
// 	bus *events.Bus,
// 	output *tview.TextView,
// ) {
// 	bus.On(ui.PrintLogEv{}, func(e any) {
// 		le, _ := e.(ui.PrintLogEv)
// 		fmt.Fprintf(output, le.Log)
// 	})
// }

type OutputWinView struct {
	TextArea *tview.TextView
}

func NewOutputWinView() *OutputWinView {
	var outputWin *tview.TextView
	outputWin = tview.NewTextView().SetChangedFunc(func() {
		outputWin.ScrollToEnd()
	})

	outputWin.SetBorder(true).SetTitle("DebugWindow")

	return &OutputWinView{
		TextArea: outputWin,
	}
}
