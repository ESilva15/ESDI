package views

import (
	"esdi/cdashdisplay"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	LayoutToolFlexID        = "layout-tool-flex"
	LayoutToolTreeID        = "layout-tool-tree"
	LayoutToolActionPagesID = "layout-tool-action-pages"
)

type windowReference struct {
	ID   int16
	Form *dom.UINode
}

func getWindowRefFromNode(node *tview.TreeNode) *windowReference {
	ref := node.GetReference()
	wRef, ok := ref.(*windowReference)

	if !ok {
		return nil
	}

	return wRef
}

func FindNodeByID(
	node *tview.TreeNode,
	id int16,
) *tview.TreeNode {
	if node == nil {
		return nil
	}

	if ref, ok := node.GetReference().(*windowReference); ok {
		if ref.ID == id {
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

func appendWindow(bus *events.Bus, doc *dom.DOM, tree *tview.TreeView, idx int16,
	win *cdashdisplay.UIWindow) {
	root := tree.GetRoot()
	if root == nil {
		bus.Emit(ui.LogEv{Log: "unable to get root of tree view"})
		return
	}

	// Create the update tool
	updateForm := windowInfoForm(bus, doc, idx, win)

	fmtTitle := fmt.Sprintf("%s [%2d]", win.Title.String(), idx)
	ref := windowReference{
		ID:   idx,
		Form: updateForm,
	}
	newWindow := tview.NewTreeNode(fmtTitle).SetReference(&ref)
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
		ev := e.(ui.WindowCreatedEv)

		bus.Emit(ui.LogEv{Log: "Received a window created event\n"})
		if tree == nil {
			bus.Emit(ui.LogEv{Log: "tree view is nil"})
			return
		}

		appendWindow(bus, doc, tree, ev.ID, &ev.Win)
	})

	bus.On(ui.WindowDestroyedEv{}, func(e any) {
		root := tree.GetRoot()
		if root == nil {
			bus.Emit(ui.LogEv{Log: "unable to get current tree node"})
		}

		node := FindNodeByID(root, e.(ui.WindowDestroyedEv).ID)
		if node != nil {
			root.RemoveChild(node)
		}

		// NODE: add a log here in case it fails so we know whats going on
		bus.Emit(ui.ForceRedraw{})
	})

	bus.On(ui.RegisterLoadedLayout{}, func(e any) {
		root := tree.GetRoot()
		if root == nil {
			return
		}

		layout := e.(ui.RegisterLoadedLayout)
		for idx, w := range layout.Layout.Windows {
			appendWindow(bus, doc, tree, idx, w)
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

func getCurNodeRef(tree *tview.TreeView) (*windowReference, error) {
	curNode := tree.GetCurrentNode()

	if curNode == nil {
		return nil, fmt.Errorf("unable to get current tree node")
	}

	winRef := getWindowRefFromNode(curNode)
	if winRef == nil {
		return nil, fmt.Errorf("unable to get window reference from tree node")
	}

	return winRef, nil
}

func layoutToolTreeViewEvCapture(bus *events.Bus, doc *dom.DOM,
	tree *tview.TreeView) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			bus.Emit(ui.ChangeFocusEv{Target: doc.GetElemByID(DeviceAPIListID)})
		}

		switch event.Rune() {
		case 'n':
			// Create a new window
			bus.Emit(ui.LogEv{Log: "calling new window form\n"})
			createNewWindowForm(bus, doc)
		case 'x':
			// Delete selected window
			bus.Emit(ui.LogEv{Log: "calling delete window\n"})
			wRef, err := getCurNodeRef(tree)
			if err != nil {
				bus.Emit(ui.LogEv{Log: " -> " + err.Error()})
				break
			}
			bus.Emit(ui.DestroyWindowEv{ID: wRef.ID})
		case 'm':
			// Go into move mode
			bus.Emit(ui.LogEv{Log: "calling window manipulation tool\n"})
			wRef, err := getCurNodeRef(tree)
			if err != nil {
				bus.Emit(ui.LogEv{Log: " -> " + err.Error()})
				break
			}
			windowManipulationTool(bus, doc, wRef.ID)
		case 'e':
			// Go into edit mode
		case 's':
			// Save the current layout
			bus.Emit(ui.SaveLayoutEv{})
		case 'l':
			// Load the layout
			bus.Emit(ui.LoadLayoutEv{})
		case 'g':
			// Go -> launches the current set up source
			streamingWindow(bus, doc)
		}

		return event
	}
}

func layoutToolTreeViewOnChange(bus *events.Bus, doc *dom.DOM) func(node *tview.TreeNode) {
	// Set focus to our new tool
	// if changeFocus {
	// 	bus.Emit(ui.ChangeFocusEv{Target: page.Self})
	// }

	return func(node *tview.TreeNode) {
		// Here we want to change the current existing form on the action pages
		nodeRef := node.GetReference()
		if nodeRef == nil {
			// its probably root I guess, thats what I'll believe
			return
		}

		nodeWinRef := nodeRef.(*windowReference)

		bus.Emit(ui.LogEv{
			Log: fmt.Sprintf("changing to -> %d | %s\n", nodeWinRef.ID, nodeWinRef.Form.ID),
		})

		pages := doc.GetElemByID(LayoutToolActionPagesID)
		AddAndShowPage(bus, doc, pages.(*tview.Pages), nodeWinRef.Form, false)
	}
}

func layoutToolTreeViewOnSelect(bus *events.Bus) func(node *tview.TreeNode) {
	return func(node *tview.TreeNode) {
		// Get the window reference which will have the ID for the form
		bus.Emit(ui.LogEv{Log: "PRESSED ENTER\n"})

		ref := node.GetReference()
		if ref == nil {
			bus.Emit(ui.LogEv{Log: "Couldn't find reference for node\n"})
			return
		}

		nodeWinRef := ref.(*windowReference)
		bus.Emit(ui.ChangeFocusEv{Target: nodeWinRef.Form.Self})
	}
}

func buildLayoutTreeComponent(bus *events.Bus, doc *dom.DOM) (*dom.UINode, error) {
	layoutTree := tview.NewTreeView()

	layoutTree.SetBorder(true).SetTitle("Layout Tree").
		SetInputCapture(layoutToolTreeViewEvCapture(bus, doc, layoutTree))

	layoutTree.SetChangedFunc(layoutToolTreeViewOnChange(bus, doc))
	layoutTree.SetSelectedFunc(layoutToolTreeViewOnSelect(bus))

	layoutTreeUINode, err := doc.NewUINode(LayoutToolTreeID,
		doc.GetElemByID(RightFlexID), layoutTree)
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

	actionPages := tview.NewPages().AddPage(EmptyPageName, emptyPage, true, true)

	actionPagesUINode, err := doc.NewUINode(LayoutToolActionPagesID,
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
	layoutToolFlexNode, err := doc.NewUINode(LayoutToolFlexID,
		doc.GetElemByID(APIToolPagesID), layoutToolFlex)
	if err != nil {
		return nil, err
	}

	layoutToolFlex.
		AddItem(layoutTreeUINode.Self, 0, 2, true).
		AddItem(actionPagesUINode.Self, 0, 5, false)

	return layoutToolFlexNode, nil
}
