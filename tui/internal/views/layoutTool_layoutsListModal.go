package views

import (
	"os"

	"github.com/rivo/tview"
)

type LayoutToolLayoutsList struct {
	Tree *tview.TreeView
}

func NewLayoutToolLayoutsList() *LayoutToolLayoutsList {
	view := &LayoutToolLayoutsList{}

	view.Tree = tview.NewTreeView()

	view.Tree.SetBorder(true).SetTitle("Layout List")
	// Inject InputCapture
	// Inject ChangedFunc
	// Inject SelectedFunc

	// Create a root element
	rootElem := tview.NewTreeNode(".")
	view.Tree.SetRoot(rootElem).SetCurrentNode(rootElem)

	return view
}

// AddEntries adds the list of files we wish the user to select from
// For now we will send just a list of []string but in the future this should
// receive (or have an alternative to receive) a directory structure
func (ltll *LayoutToolLayoutsList) AddEntries(entries []os.DirEntry) {
	root := ltll.Tree.GetRoot()
	if root == nil {
		// NOTE: shouldn't happen at all
		return
	}

	for _, e := range entries {
		node := tview.NewTreeNode(e.Name()).
			SetReference(e.Name())
		root.AddChild(node)
	}
}
