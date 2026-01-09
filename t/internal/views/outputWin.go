package views

import (
	"esdi/t/internal/dom"
	"esdi/t/internal/ui"

	"github.com/rivo/tview"
)

func BuildOutputWindow(root *dom.DOM, ctx *ui.UIContext) error {
	var outputWin *tview.TextView
	outputWin = tview.NewTextView().SetChangedFunc(func() {
		ctx.Redraw()
		outputWin.ScrollToEnd()
	})

	outputWin.SetBorder(true).SetTitle("DebugWindow")
	_, err := root.NewUINode("output-window", nil, outputWin)
	if err != nil {
		return err
	}

	return nil
}
