// Package dom
package dom

import (
	"github.com/rivo/tview"
)

type UINode struct {
	ID       string
	Self     tview.Primitive
	Parent   tview.Primitive
	Children map[string]*UINode
}

func newUINode(ID string, parent, root tview.Primitive) *UINode {
	return &UINode{
		ID:       ID,
		Self:     root,
		Children: make(map[string]*UINode),
	}
}

func (l *UINode) AppendItem(item *UINode) error {
	// Its not necessary to verify if this ID is already a child of this Node
	// because we do that at the UI level
	l.Children[item.ID] = item
	return nil
}
