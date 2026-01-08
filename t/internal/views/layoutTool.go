package views

import (
	"esdi/t/internal/dom"

	"github.com/rivo/tview"
)

func buildLayoutToolUINode(root *dom.DOM) (*dom.UINode, error) {
	// Want a TreeView at the top
	layoutTree := tview.NewTreeView()
	layoutTree.SetBorder(true).SetTitle("Layout Tree")
	_, err := root.NewUINode("layout-tree",
		root.GetElemByID("right-flex"), layoutTree)
	if err != nil {
		return nil, err
	}

	// Want pages below to run actions
	actionPages := tview.NewPages()
	actionPages.SetBorder(true).SetTitle("Action Page")
	_, err = root.NewUINode("layout-action-pages",
		root.GetElemByID("right-flex"), actionPages)
	if err != nil {
		return nil, err
	}

	layoutToolFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	layoutToolFlexNode, err := root.NewUINode("layout-tool-flex", root.GetElemByID("api-tool-pages"),
		layoutToolFlex)
	if err != nil {
		return nil, err
	}

	return layoutToolFlexNode, nil
}
