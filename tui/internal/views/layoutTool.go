package views

import (
	"fmt"
	"log/slog"

	"esdi/cdashdisplay"
	tviewh "esdi/tui/internal/tview_helpers"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	LayoutToolTreeViewTitle = "Layout Tree"
	LayoutToolFlexID        = "layout-tool-flex"
	LayoutToolTreeID        = "layout-tool-tree"
	LayoutToolActionPagesID = "layout-tool-action-pages"
	LayoutToolNewWindowID   = "new-window-action-form"
)

type LayoutTreeView struct {
	Tree         *tview.TreeView
	InputCapture func(*tcell.EventKey) *tcell.EventKey
}

func NewLayoutTreeView() *LayoutTreeView {
	view := &LayoutTreeView{
		InputCapture: blankInputCapture,
	}

	view.Tree = tview.NewTreeView()

	view.Tree.SetBorder(true).SetTitle(LayoutToolTreeViewTitle)
	view.Tree.SetInputCapture(view.InputCapture)
	// Inject ChangedFunc
	// Inject SelectedFunc

	// Create a root element
	rootElem := tview.NewTreeNode(".")
	view.Tree.SetRoot(rootElem).SetCurrentNode(rootElem)

	return view
}

func (lt *LayoutTreeView) AddWindow(win *cdashdisplay.DesktopUIWindow) error {
	root := lt.Tree.GetRoot()
	if root == nil {
		return fmt.Errorf("unable to get root of treeview")
	}

	// Create this new node
	newWindow := tview.NewTreeNode(
		windowInfoPageTitle(win.UIData.IDX, win.Title.String()),
	).SetReference(win.UIData.IDX)

	root.AddChild(newWindow)

	return nil
}

type LayoutToolActionView struct {
	Pages            *tview.Pages
	CreateWindowView *CreateWindowFormView
	LayoutSelection  *LayoutToolLayoutsList
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

func (ltv *LayoutToolView) WindowCreatedSuccessfuly(
	lcLogger *slog.Logger, win *cdashdisplay.DesktopUIWindow,
) error {
	lcLogger.Debug(fmt.Sprintf("creating window form view for: %+v", win.Title))
	updateWindowForm := NewWindowFormView(win)

	// Update the pages ID
	lcLogger.Debug(fmt.Sprintf("removing page: %+v", LayoutToolNewWindowID))
	ltv.LayoutActions.Pages.RemovePage(LayoutToolNewWindowID)
	lcLogger.Debug(fmt.Sprintf("adding page: %+v", LayoutToolNewWindowID))
	AddAndShowPage(
		ltv.LayoutActions.Pages,
		windowInfoPageID(win.UIData.IDX),
		updateWindowForm.Form.Form,
	)

	// Then set the CreateWindowView pointer to nil
	lcLogger.Debug(fmt.Sprintf("setting create window pointer to nil: %+v", LayoutToolNewWindowID))
	ltv.LayoutActions.CreateWindowView = nil

	// Add this form thing to the quick access map
	lcLogger.Debug(fmt.Sprintf("adding form to quick access map: %+v", LayoutToolNewWindowID))
	ltv.FormQuickAccess[win.UIData.IDX] = updateWindowForm

	// Add it to the tree view
	lcLogger.Debug(fmt.Sprintf("adding it to tree view: %+v", LayoutToolNewWindowID))
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

func (ltv *LayoutToolView) ShowLayoutList() {
	AddAndShowPage(
		ltv.LayoutActions.Pages,
		"available-layouts-list",
		ltv.LayoutActions.LayoutSelection.Tree,
	)
}

func (ltv *LayoutToolView) DeleteLayoutList() {
	ltv.LayoutActions.Pages.RemovePage("available-layouts-list")
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
	delete(ltv.FormQuickAccess, node.GetReference().(int16))
	// Did we delete everything ???
}

func (ltv *LayoutToolView) UpdateFormView(idx int16, win *cdashdisplay.DesktopUIWindow) {
	// Update the form view
	ltv.FormQuickAccess[idx].SetValues(win)

	// Update the treeview with the new title
	root := ltv.LayoutTree.Tree.GetRoot()
	if root == nil {
		// NOTE: excuse me!?
		return
	}

	node := tviewh.FindNodeByID(root, idx)
	if node == nil {
		return
	}

	node.SetText(windowInfoPageTitle(idx, win.Title.String()))
}
