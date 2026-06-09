package tviewhelpers

import "github.com/rivo/tview"

func FindNodeByID(node *tview.TreeNode, id int16) *tview.TreeNode {
	if node == nil {
		return nil
	}

	if ref, ok := node.GetReference().(int16); ok {
		if ref == id {
			return node
		}
	}

	for _, child := range node.GetChildren() {
		if found := FindNodeByID(child, id); found != nil {
			return found
		}
	}

	return nil
}
