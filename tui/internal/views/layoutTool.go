package views

import (
	"esdi/tui/internal/models"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	LayoutToolFlexID        = "layout-tool-flex"
	LayoutToolTreeID        = "layout-tool-tree"
	LayoutToolActionPagesID = "layout-tool-action-pages"
	LayoutToolNewWindowID   = "new-window-action-form"
)

// type windowReference struct {
// 	ID   int16
// 	Form *dom.UINode
// }

// func getWindowRefFromNode(node *tview.TreeNode) *windowReference {
// 	ref := node.GetReference()
// 	wRef, ok := ref.(*windowReference)
//
// 	if !ok {
// 		return nil
// 	}
//
// 	return wRef
// }

func FindNodeByID(node *tview.TreeNode, id int16) *tview.TreeNode {
	if node == nil {
		return nil
	}

	if ref, ok := node.GetReference().(*models.UIWindow); ok {
		if ref.IDX == id {
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

type LayoutTreeView struct {
	Tree         *tview.TreeView
	InputCapture func(*tcell.EventKey) *tcell.EventKey
	// OnChange     func(node *tview.TreeNode)
	// OnSelect     func(node *tview.TreeNode)
}

func NewLayoutTreeView() *LayoutTreeView {
	view := &LayoutTreeView{
		InputCapture: blankInputCapture,
		// OnChange:     blankTreeViewOnChange,
		// OnSelect:     blankTreeViewOnChange,
	}

	view.Tree = tview.NewTreeView()

	view.Tree.SetBorder(true).SetTitle("Layout Tree")
	view.Tree.SetInputCapture(view.InputCapture)
	// Inject ChangedFunc
	// Inject SelectedFunc

	// Create a root element
	rootElem := tview.NewTreeNode(".")
	view.Tree.SetRoot(rootElem).SetCurrentNode(rootElem)

	return view
}

func (lt *LayoutTreeView) AddWindow(win *models.UIWindow) error {
	root := lt.Tree.GetRoot()
	if root == nil {
		return fmt.Errorf("unable to get root of treeview")
	}

	// Create this new node
	newWindow := tview.NewTreeNode(
		windowInfoPageTitle(win.IDX, win.Window.Title.String()),
	).SetReference(win)

	root.AddChild(newWindow)

	return nil
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
	Flex            *tview.Flex
	FormQuickAccess map[int16]*WindowFormView
	LayoutTree      *LayoutTreeView
	LayoutActions   *LayoutToolActionView
}

func NewLayoutToolView() *LayoutToolView {
	view := &LayoutToolView{
		FormQuickAccess: make(map[int16]*WindowFormView),
	}

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

func (ltv *LayoutToolView) WindowCreatedSuccessfuly(win *models.UIWindow) error {
	updateWindowForm := NewWindowFormView(win)

	// Update the pages ID
	ltv.LayoutActions.Pages.RemovePage(LayoutToolNewWindowID)
	AddAndShowPage(
		ltv.LayoutActions.Pages,
		windowInfoPageID(win.IDX),
		updateWindowForm.Form.Form,
	)

	// Then set the CreateWindowView pointer to nil
	ltv.LayoutActions.CreateWindowView = nil

	// Add this form thing to the quick access map
	ltv.FormQuickAccess[win.IDX] = updateWindowForm

	// Add it to the tree view
	return ltv.LayoutTree.AddWindow(win)
}

func (ltv *LayoutToolView) ShowCreateWindowForm() {
	AddAndShowPage(
		ltv.LayoutActions.Pages,
		LayoutToolNewWindowID,
		ltv.LayoutActions.CreateWindowView.Form.Form,
	)
}

func (ltv *LayoutToolView) ShowWindowFormByID(idx int16) {
	AddAndShowPage(
		ltv.LayoutActions.Pages,
		windowInfoPageID(idx),
		nil,
	)
}

func (ltv *LayoutToolView) DeleteWindowByNode(node *tview.TreeNode) {
	root := ltv.LayoutTree.Tree.GetRoot()
	if root == nil {
		// We should do something I guess
		return
	}

	root.RemoveChild(node)

	// We have to delete it from the map
	// This isn't very safe now is it?
	delete(ltv.FormQuickAccess, node.GetReference().(*models.UIWindow).IDX)
	// Did we delete everything ???
}

func (ltv *LayoutToolView) UpdateFormView(win *models.UIWindow) {
	// Update the form view
	ltv.FormQuickAccess[win.IDX].SetValues(win)

	// Update the treeview with the new title
	root := ltv.LayoutTree.Tree.GetRoot()
	if root == nil {
		// NOTE: excuse me!?
		return
	}

	node := FindNodeByID(root, win.IDX)
	if node == nil {
		return
	}

	node.SetText(windowInfoPageTitle(win.IDX, win.Window.Title.String()))
}
