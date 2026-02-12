package views

import (
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	layoutToolFlexID        = "layout-tool-flex"
	layoutToolTreeID        = "layout-tool-tree"
	layoutToolActionPagesID = "layout-tool-action-pages"
)

type windowReference struct {
	ID int16
}

func getWindowRefFromNode(node *tview.TreeNode) *windowReference {
	ref := node.GetReference()
	wRef, ok := ref.(windowReference)

	if !ok {
		return nil
	}

	return &wRef
}

func FindNodeByReference(
	node *tview.TreeNode,
	want any,
) *tview.TreeNode {
	if node == nil {
		return nil
	}

	if node.GetReference() == want {
		return node
	}

	for _, child := range node.GetChildren() {
		if found := FindNodeByReference(child, want); found != nil {
			return found
		}
	}

	return nil
}

func appendWindow(bus *events.Bus, tree *tview.TreeView, idx int16, title string) {
	root := tree.GetRoot()
	if root == nil {
		bus.Emit(ui.LogEv{Log: "unable to get root of tree view"})
		return
	}

	fmtTitle := fmt.Sprintf("%s [%2d]", title, idx)
	ref := windowReference{
		ID: idx,
	}
	newWindow := tview.NewTreeNode(fmtTitle).SetReference(ref)
	root.AddChild(newWindow)
}

func BindWindowEvents(
	bus *events.Bus,
	doc *dom.DOM,
	tree *tview.TreeView,
) {
	// I recon I have to change this for some type os event system that
	// triggers directly on the UINodes I want them to be triggered on

	bus.On(ui.WindowCreatedEv{}, func(e any) {
		go func() {
			win := e.(ui.WindowCreatedEv)

			bus.Emit(ui.LogEv{Log: "Received a window created event\n"})
			if tree == nil {
				bus.Emit(ui.LogEv{Log: "tree view is nil"})
				return
			}

			appendWindow(bus, tree, win.ID, win.Title)
		}()
	})

	bus.On(ui.WindowDestroyedEv{}, func(e any) {
		go func() {
			root := tree.GetRoot()
			if root == nil {
				bus.Emit(ui.LogEv{Log: "unable to get current tree node"})
			}

			node := FindNodeByReference(root, e.(ui.WindowDestroyedEv).ID)
			if node != nil {
				root.RemoveChild(node)
			}

			// NODE: add a log here in case it fails so we know whats going on
			bus.Emit(ui.ForceRedraw{})
		}()
	})

	bus.On(ui.RegisterLoadedLayout{}, func(e any) {
		root := tree.GetRoot()
		if root == nil {
			return
		}

		layout := e.(ui.RegisterLoadedLayout)
		for idx, w := range layout.Layout.Windows {
			appendWindow(bus, tree, idx, w.Title.String())
		}

		// bus.Emit(ui.ForceRedraw{})
	})

	bus.On(ui.ErrorCreateWindowEv{}, func(e any) {
		go func() {
			bus.Emit(ui.LogEv{
				Log: fmt.Sprintf(
					"Error performing action: %s\n", e.(ui.ErrorCreateWindowEv).Error.Error(),
				)})
		}()
	})
}

func layoutToolTreeViewEvCapture(bus *events.Bus, doc *dom.DOM,
	tree *tview.TreeView) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			bus.Emit(ui.ChangeFocusEv{Target: doc.GetElemByID(deviceAPIListID)})
		}

		switch event.Rune() {
		case 'n':
			// Create a new window
			bus.Emit(ui.LogEv{Log: "calling new window form\n"})
			createNewWindowForm(bus, doc)
		case 'x':
			// Delete selected window
			bus.Emit(ui.LogEv{Log: "calling delete window\n"})

			curNode := tree.GetCurrentNode()
			if curNode == nil {
				bus.Emit(ui.LogEv{Log: "unable to get current tree node"})
				break
			}

			winRef := getWindowRefFromNode(curNode)
			if winRef == nil {
				// Whatever, do something better here
				break
			}

			bus.Emit(ui.DestroyWindowEv{ID: winRef.ID})
		case 'm':
			// Go into move mode
			curNode := tree.GetCurrentNode()
			if curNode == nil {
				bus.Emit(ui.LogEv{Log: "unable to get current tree node"})
				break
			}

			winRef := getWindowRefFromNode(curNode)
			if winRef == nil {
				// Whatever, do something better here
				break
			}

			windowManipulationTool(bus, doc, winRef.ID)
		case 'e':
			// Go into edit mode
		case 's':
			// Save the current layout
			bus.Emit(ui.SaveLayoutEv{})
		case 'l':
			// Load the layout
			bus.Emit(ui.LoadLayoutEv{})
		}

		return event
	}
}

func layoutToolUIOnSelect(bus *events.Bus, doc *dom.DOM) {
	bus.Emit(ui.LogEv{Log: "Opening layout tool UI\n"})
	var err error

	// Get the api pages
	apiToolPages := doc.GetElemByID(apiToolPagesID).(*tview.Pages)
	if apiToolPages == nil {
		bus.Emit(ui.LogEv{Log: "  Failed to retrive apiToolPages UI\n"})
		return
	}

	bus.Emit(ui.LogEv{
		Log: fmt.Sprintf("  Checking if `%s` UINode already exists\n", layoutToolFlexID),
	})
	layoutToolUINode := doc.GetNodeByID(layoutToolFlexID)
	if layoutToolUINode == nil {
		layoutToolUINode, err = buildLayoutFlexComponent(bus, doc)
		if err != nil {
			bus.Emit(ui.LogEv{
				Log: fmt.Sprintf("      Failed to build layout tool UI: %s\n", err.Error()),
			})
		}
	}

	AddAndShowPage(bus, doc, apiToolPages, layoutToolUINode)
}

func layoutToolTreeViewOnChange() func(node *tview.TreeNode) {
	return func(node *tview.TreeNode) {
		// Here we want to change the current existing form on the action pages
	}
}

func buildLayoutTreeComponent(bus *events.Bus, doc *dom.DOM) (*dom.UINode, error) {
	layoutTree := tview.NewTreeView()

	layoutTree.SetBorder(true).SetTitle("Layout Tree").
		SetInputCapture(layoutToolTreeViewEvCapture(bus, doc, layoutTree))

	layoutTree.SetChangedFunc(layoutToolTreeViewOnChange())

	layoutTreeUINode, err := doc.NewUINode(layoutToolTreeID,
		doc.GetElemByID(rightFlexID), layoutTree)
	if err != nil {
		return nil, err
	}

	// Bind the events for the treeview
	BindWindowEvents(bus, doc, layoutTreeUINode.Self.(*tview.TreeView))

	// Create a root element
	rootElem := tview.NewTreeNode(".")
	layoutTree.SetRoot(rootElem).SetCurrentNode(rootElem)

	return layoutTreeUINode, nil
}

func buildLayoutActionPagesComponent(bus *events.Bus, doc *dom.DOM) (*dom.UINode, error) {
	// Create an empty page for it
	emptyPage := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			bus.Emit(ui.RedrawEv{})
		})
	emptyPage.SetBorder(true).SetTitle("-- Tool Area --")
	fmt.Fprintf(emptyPage, "No tool selected")

	actionPages := tview.NewPages().AddPage(emptyPageName, emptyPage, true, true)

	actionPagesUINode, err := doc.NewUINode(layoutToolActionPagesID,
		doc.GetElemByID("right-flex"), actionPages)
	if err != nil {
		return nil, err
	}

	return actionPagesUINode, nil
}

func buildLayoutFlexComponent(bus *events.Bus, doc *dom.DOM) (*dom.UINode, error) {
	// Want a TreeView at the top
	layoutTreeUINode, err := buildLayoutTreeComponent(bus, doc)
	if err != nil {
		return nil, err
	}

	// Want pages below to run actions
	actionPagesUINode, err := buildLayoutActionPagesComponent(bus, doc)
	if err != nil {
		return nil, err
	}

	layoutToolFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	layoutToolFlexNode, err := doc.NewUINode(layoutToolFlexID,
		doc.GetElemByID(apiToolPagesID), layoutToolFlex)
	if err != nil {
		return nil, err
	}

	layoutToolFlex.
		AddItem(layoutTreeUINode.Self, 0, 2, true).
		AddItem(actionPagesUINode.Self, 0, 5, false)

	return layoutToolFlexNode, nil
}
