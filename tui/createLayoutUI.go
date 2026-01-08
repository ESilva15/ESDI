package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type cDashWinLayout struct {
	tree *tview.TreeView
}

func createNewWindowForm(t *tui, dp *devicePane, tree *tview.TreeView, f *tview.Pages) {
	x0 := tview.NewInputField().SetLabel("x")
	y0 := tview.NewInputField().SetLabel("y")
	width := tview.NewInputField().SetLabel("width")
	height := tview.NewInputField().SetLabel("height")
	title := tview.NewInputField().SetLabel("title")

	form := tview.NewForm().
		AddFormItem(x0).
		AddFormItem(y0).
		AddFormItem(width).
		AddFormItem(height).
		AddFormItem(title).
		AddButton("Create", func() {
			createWindow(
				t,
				x0.GetText(),
				y0.GetText(),
				width.GetText(),
				height.GetText(),
				title.GetText(),
				tree,
			)
		})

	form.SetBorder(true).
		SetTitle("new window form").
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Key() {
			case tcell.KeyEscape:
				f.RemovePage("new-window-form")
				t.app.SetFocus(tree)
			}

			return ev
		})

	f.RemovePage("new-window-form")
	f.AddPage("new-window-form", form, true, true)

	t.app.SetFocus(form)
}

func createLayoutUIFormKeybinds(t *tui, dp *devicePane,
	tree *tview.TreeView, f *tview.Pages) func(*tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEscape:
			t.app.SetFocus(dp.apiPane)
		}

		switch ev.Rune() {
		case 'N':
			// Create a new screen
		case 'n':
			// Create a new window
			t.log("calling new window form\n")
			createNewWindowForm(t, dp, tree, f)
		case 'x':
			// Delete selected window
		case 'm':
			// Go into move mode
		case 'e':
			// Go into edit mode
		}

		return ev
	}
}

func createLayoutUI(t *tui, dp *devicePane) {
	layoutTree := tview.NewTreeView()
	actionPages := tview.NewPages()

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(layoutTree, 0, 1, true).
		AddItem(actionPages, 0, 2, true)

	root := tview.NewTreeNode("Root").SetColor(tcell.ColorRed)
	layoutTree.SetRoot(root).SetCurrentNode(root)
	layoutTree.
		SetBorder(true).
		SetTitle("Tree").
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(createLayoutUIFormKeybinds(t, dp, layoutTree, actionPages))

	dp.actionPane.RemovePage("action")
	dp.actionPane.AddPage("action", flex, true, true)

	t.app.SetFocus(layoutTree)
}
