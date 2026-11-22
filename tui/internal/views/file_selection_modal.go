package views

import (
	"github.com/rivo/tview"
)

// TODO: maybe implement telescope fuzzy finder here?
// It could have extra functionalities out of the box I guess
// TODO: Should the controller build the *tview.TreeNodes? Or should the view
// actually parse the []os.DirEntry entires itself?
func FileSelectionModal(
	root *tview.TreeNode,
	onSelect func(node *tview.TreeNode),
) *tview.TreeView {
	tree := tview.NewTreeView()

	tree.SetSelectedFunc(onSelect)

	tree.SetBorder(true).SetTitle("File Picker")

	tree.SetRoot(root)
	tree.SetCurrentNode(root)

	return tree
}
