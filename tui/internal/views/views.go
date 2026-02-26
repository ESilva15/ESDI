package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	blankInputCapture           = func(ev *tcell.EventKey) *tcell.EventKey { return ev }
	blankTreeViewOnChange       = func(node *tview.TreeNode) {}
	blankDropdownOptionCallback = func() {}
)
