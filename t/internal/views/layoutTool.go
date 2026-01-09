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
	layoutToolFlexID = "layout-tool-flex"
)

func BindWindowEvents(
	ctx *ui.UIContext,
	bus *events.Bus,
	tree *tview.TreeView,
) {
	bus.On(events.WindowCreated{}, func(e any) {
		ctx.Log("Received a window created event")
	})
}

func layoutToolTreeViewEvents(root *dom.DOM, ctx *ui.UIContext,
	wc *controllers.WindowingController) func(
	event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			ctx.ChangeFocus(root.GetElemByID(deviceAPIListID))
		}

		switch event.Rune() {
		case 'N':
			// Create a new screen
		case 'n':
			// Create a new window
			ctx.Log("calling new window form\n")
			createNewWindowForm(root, ctx, wc)
		case 'x':
			// Delete selected window
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
		layoutToolUINode, err = buildLayoutToolUINode(root, ctx, wc)
		if err != nil {
			ctx.Log("      Failed to build layout tool UI: %s\n", err.Error())
		}
	}

	AddAndShowPage(root, ctx, apiToolPages, layoutToolUINode)
}

func buildLayoutToolUINode(root *dom.DOM, ctx *ui.UIContext,
	wc *controllers.WindowingController) (*dom.UINode, error) {
	// Want a TreeView at the top
	layoutTree := tview.NewTreeView().
		SetInputCapture(layoutToolTreeViewEvents(root, ctx, wc))
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
	layoutToolFlexNode, err := root.NewUINode(layoutToolFlexID, root.GetElemByID("api-tool-pages"),
		layoutToolFlex)
	if err != nil {
		return nil, err
	}

	layoutToolFlex.
		AddItem(layoutTree, 0, 2, true).
		AddItem(actionPages, 0, 5, false)

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
		formNode, err = root.NewUINode("new-window-form", root.GetElemByID("layout-action-pages"),
			form)
		if err != nil {
			panic("failed to create UI node for the new window form: " + err.Error())
		}
	}

	AddAndShowPage(root, ctx, root.GetElemByID("layout-action-pages").(*tview.Pages),
		formNode)
}
