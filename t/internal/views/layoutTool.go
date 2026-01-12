package views

import (
	"esdi/t/internal/controllers"
	"esdi/t/internal/dom"
	"esdi/t/internal/events"
	"esdi/t/internal/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	layoutToolFlexID        = "layout-tool-flex"
	layoutToolTreeID        = "layout-tool-tree"
	layoutToolActionPagesID = "layout-tool-action-pages"
)

func BindWindowEvents(
	ctx *ui.UIContext,
	bus *events.Bus,
	tree *tview.TreeView,
) {
	// I recon I have to change this for some type os event system that
	// triggers directly on the UINodes I want them to be triggered on
	bus.On(events.WindowCreated{}, func(e any) {
		ctx.Log("Received a window created event\n")
		if tree == nil {
			ctx.Log("tree view is nil")
			return
		}

		root := tree.GetRoot()
		if root == nil {
			ctx.Log("unable to get root of tree view")
			return
		}

		newWindow := tview.NewTreeNode(e.(events.WindowCreated).Window.Title)
		root.AddChild(newWindow)
	})
	bus.On(events.Error{}, func(e any) {
		ctx.Log("Error performing action: %s\n", e.(events.Error).Error.Error())
	})
}

func layoutToolTreeViewEvents(root *dom.DOM, ctx *ui.UIContext,
	tree *tview.TreeView, wc *controllers.WindowingController) func(
	event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			ctx.ChangeFocus(root.GetElemByID(deviceAPIListID))
		}

		switch event.Rune() {
		case 'N':
			// Create a new screen
			// I think I will leave this one for later
		case 'n':
			// Create a new window
			ctx.Log("calling new window form\n")
			createNewWindowForm(root, ctx, wc)
		case 'x':
			// Delete selected window
			ctx.Log("calling delete window\n")
			curNode := tree.GetCurrentNode()
			if curNode == nil {
				ctx.Log("unable to get current tree node")
				break
			}

			root := tree.GetRoot()
			if root == nil {
				ctx.Log("unable to get current tree node")
				break
			}

			root.RemoveChild(curNode)
		case 'm':
			// Go into move mode
		case 'e':
			// Go into edit mode
		}

		return event
	}
}

func layoutToolUIOnSelect(root *dom.DOM, ctx *ui.UIContext, wc *controllers.WindowingController) {
	ctx.Log("Opening layout tool UI\n")
	var err error

	// Get the api pages
	apiToolPages := root.GetElemByID(apiToolPagesID).(*tview.Pages)
	if apiToolPages == nil {
		ctx.Log("  Failed to retrive apiToolPages UI\n")
		return
	}

	ctx.Log("  Checking if `%s` UINode already exists\n", layoutToolFlexID)
	layoutToolUINode := root.GetNodeByID(layoutToolFlexID)
	if layoutToolUINode == nil {
		layoutToolUINode, err = buildLayoutFlexComponent(root, ctx, wc)
		if err != nil {
			ctx.Log("      Failed to build layout tool UI: %s\n", err.Error())
		}
	}

	AddAndShowPage(root, ctx, apiToolPages, layoutToolUINode)
}

func buildLayoutTreeComponent(doc *dom.DOM, ctx *ui.UIContext,
	wc *controllers.WindowingController) (*dom.UINode, error) {

	layoutTree := tview.NewTreeView()

	layoutTree.SetBorder(true).SetTitle("Layout Tree").
		SetInputCapture(layoutToolTreeViewEvents(doc, ctx, layoutTree, wc))

	layoutTreeUINode, err := doc.NewUINode(layoutToolTreeID,
		doc.GetElemByID(rightFlexID), layoutTree)
	if err != nil {
		return nil, err
	}

	// Bind the events for the treeview
	BindWindowEvents(ctx, wc.Events, layoutTreeUINode.Self.(*tview.TreeView))

	// Create a root element
	rootElem := tview.NewTreeNode(".")
	layoutTree.SetRoot(rootElem).SetCurrentNode(rootElem)

	return layoutTreeUINode, nil
}

func buildLayoutActionPagesComponent(doc *dom.DOM, _ *ui.UIContext,
	_ *controllers.WindowingController) (*dom.UINode, error) {

	actionPages := tview.NewPages()

	actionPages.SetBorder(true).SetTitle("Action Page")

	actionPagesUINode, err := doc.NewUINode(layoutToolActionPagesID,
		doc.GetElemByID("right-flex"), actionPages)
	if err != nil {
		return nil, err
	}

	return actionPagesUINode, nil
}

func buildLayoutFlexComponent(root *dom.DOM, ctx *ui.UIContext,
	wc *controllers.WindowingController) (*dom.UINode, error) {
	// Want a TreeView at the top
	layoutTreeUINode, err := buildLayoutTreeComponent(root, ctx, wc)
	if err != nil {
		return nil, err
	}

	// Want pages below to run actions
	actionPagesUINode, err := buildLayoutActionPagesComponent(root, ctx, wc)
	if err != nil {
		return nil, err
	}

	layoutToolFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	layoutToolFlexNode, err := root.NewUINode(layoutToolFlexID,
		root.GetElemByID(apiToolPagesID), layoutToolFlex)
	if err != nil {
		return nil, err
	}

	layoutToolFlex.
		AddItem(layoutTreeUINode.Self, 0, 2, true).
		AddItem(actionPagesUINode.Self, 0, 5, false)

	return layoutToolFlexNode, nil
}

// Layout tool actions
func createNewWindowForm(root *dom.DOM, ctx *ui.UIContext,
	wc *controllers.WindowingController) {
	var err error

	x0 := tview.NewInputField().SetLabel("x")
	y0 := tview.NewInputField().SetLabel("y")
	width := tview.NewInputField().SetLabel("width")
	height := tview.NewInputField().SetLabel("height")
	title := tview.NewInputField().SetLabel("title")

	form := tview.NewForm().
		AddFormItem(x0).
		AddFormItem(y0).
		AddFormItem(width).
		AddFormItem(height).
		AddFormItem(title).
		AddButton("Create", func() {
			wc.CreateWindow(
				x0.GetText(),
				y0.GetText(),
				width.GetText(),
				height.GetText(),
				title.GetText(),
			)
		})

	form.SetBorder(true).
		SetTitle("new window form").
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Key() {
			case tcell.KeyEscape:
				ctx.ChangeFocus(root.GetElemByID(layoutToolFlexID))
			}

			return ev
		})

	var formNode *dom.UINode
	formNode = root.GetNodeByID("new-window-form")
	if formNode == nil {
		formNode, err = root.NewUINode("new-window-form",
			root.GetElemByID(layoutToolActionPagesID), form)
		if err != nil {
			panic("failed to create UI node for the new window form: " + err.Error())
		}
	}

	AddAndShowPage(root, ctx, root.GetElemByID(layoutToolActionPagesID).(*tview.Pages),
		formNode)
}
