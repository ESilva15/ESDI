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
	// root := tree.GetRoot()
	// if root == nil {
	// 	bus.Emit(ui.LogEv{Log: "unable to get root of tree view"})
	// 	return
	// }
	//
	// // Create the update tool
	// updateForm := windowInfoForm(bus, doc, idx, win)
	//
	// fmtTitle := fmt.Sprintf("%s [%2d]", win.Title.String(), idx)
	// ref := windowReference{
	// 	ID:   idx,
	// 	Form: updateForm,
	// }
	// newWindow := tview.NewTreeNode(fmtTitle).SetReference(&ref)
	// root.AddChild(newWindow)
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

func layoutToolTreeViewOnChange(bus *events.Bus, doc *dom.DOM) func(node *tview.TreeNode) {
	// Set focus to our new tool
	// if changeFocus {
	// 	bus.Emit(ui.ChangeFocusEv{Target: page.Self})
	// }

	return func(node *tview.TreeNode) {
		// // Here we want to change the current existing form on the action pages
		// nodeRef := node.GetReference()
		// if nodeRef == nil {
		// 	// its probably root I guess, thats what I'll believe
		// 	return
		// }
		//
		// nodeWinRef := nodeRef.(*windowReference)
		//
		// bus.Emit(ui.LogEv{
		// 	Log: fmt.Sprintf("changing to -> %d | %s\n", nodeWinRef.ID, nodeWinRef.Form.ID),
		// })
		//
		// pages := doc.GetElemByID(LayoutToolActionPagesID)
		// AddAndShowPage(pages.(*tview.Pages), nodeWinRef.Form, false)
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

type LayoutTreeView struct {
	Tree         *tview.TreeView
	InputCapture func(*tcell.EventKey) *tcell.EventKey
	OnChange     func(node *tview.TreeNode)
	OnSelect     func(node *tview.TreeNode)
}

func NewLayoutTreeView() *LayoutTreeView {
	view := &LayoutTreeView{
		InputCapture: blankInputCapture,
		OnChange:     blankTreeViewOnChange,
		OnSelect:     blankTreeViewOnChange,
	}

	view.Tree = tview.NewTreeView()

	view.Tree.SetBorder(true).SetTitle("Layout Tree")
	view.Tree.SetInputCapture(view.InputCapture)
	view.Tree.SetChangedFunc(view.OnChange)
	view.Tree.SetSelectedFunc(view.OnSelect)

	// view.Tree.SetChangedFunc(layoutToolTreeViewOnChange(bus, doc))
	// view.Tree.SetSelectedFunc(layoutToolTreeViewOnSelect(bus))

	// Bind the events for the treeview
	// BindWindowEvents(bus, doc, layoutTreeUINode.Self.(*tview.TreeView))

	// Create a root element
	rootElem := tview.NewTreeNode(".")
	view.Tree.SetRoot(rootElem).SetCurrentNode(rootElem)

	return view
}

type LayoutToolActionView struct {
	Pages            *tview.Pages
	CreateWindowView *CreateWindowFormView
}

func NewLayoutToolActionView() *LayoutToolActionView {
	view := &LayoutToolActionView{
		CreateWindowView: nil,
	}

	// Create an empty page for it
	emptyPage := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {}) // Need to hook here

	emptyPage.SetBorder(true).SetTitle("-- Tool Area --")
	fmt.Fprintf(emptyPage, "No tool selected")

	view.Pages = tview.NewPages().AddPage(EmptyPageName, emptyPage, true, true)

	return view
}

type LayoutToolView struct {
	Flex          *tview.Flex
	LayoutTree    *LayoutTreeView
	LayoutActions *LayoutToolActionView
}

func NewLayoutToolView() *LayoutToolView {
	view := &LayoutToolView{}

	// Want a TreeView at the top
	view.LayoutTree = NewLayoutTreeView()

	// Want pages below to run actions
	view.LayoutActions = NewLayoutToolActionView()

	// Create the flex
	view.Flex = tview.NewFlex().SetDirection(tview.FlexRow)

	view.Flex.
		AddItem(view.LayoutTree.Tree, 0, 2, true).
		AddItem(view.LayoutActions.Pages, 0, 5, false)

	return view
}

func (ltv *LayoutToolView) AddWindow() {
}
