package tui

import "github.com/rivo/tview"

func createWindow(t *tui, x, y, width, height, title string, tree *tview.TreeView) {
	t.log("x     : %s\n", x)
	t.log("y     : %s\n", y)
	t.log("width : %s\n", width)
	t.log("height: %s\n", height)
	t.log("title : %s\n", title)

	newWindow := tview.NewTreeNode(title)
}
