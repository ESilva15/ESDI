package views

import (
	"github.com/rivo/tview"
)

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
