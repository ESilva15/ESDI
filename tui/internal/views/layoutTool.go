package views

import (
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/models"
	"esdi/tui/internal/ui"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	layoutToolFlexID        = "layout-tool-flex"
	layoutToolTreeID        = "layout-tool-tree"
	layoutToolActionPagesID = "layout-tool-action-pages"
)

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

func BindWindowEvents(
	bus *events.Bus,
	doc *dom.DOM,
	tree *tview.TreeView,
) {
	// I recon I have to change this for some type os event system that
	// triggers directly on the UINodes I want them to be triggered on
	bus.On(ui.WindowCreatedEv{}, func(e any) {
		bus.Emit(ui.LogEv{Log: "Received a window created event\n"})
		if tree == nil {
			bus.Emit(ui.LogEv{Log: "tree view is nil"})
			return
		}

		root := tree.GetRoot()
		if root == nil {
			bus.Emit(ui.LogEv{Log: "unable to get root of tree view"})
			return
		}

		newWindow := tview.NewTreeNode(e.(ui.WindowCreatedEv).Title).
			SetReference(e.(ui.WindowCreatedEv).ID)
		root.AddChild(newWindow)
	})

	bus.On(ui.WindowDestroyedEv{}, func(e any) {
		root := tree.GetRoot()
		if root == nil {
			bus.Emit(ui.LogEv{Log: "unable to get current tree node"})
		}

		node := FindNodeByReference(root, e.(ui.WindowDestroyedEv).ID)
		if node != nil {
			root.RemoveChild(node)
		}
		// NODE: add a log here in case it fails so we know whats going on
	})

	bus.On(ui.ErrorCreateWindowEv{}, func(e any) {
		bus.Emit(ui.LogEv{
			Log: fmt.Sprintf(
				"Error performing action: %s\n", e.(ui.ErrorCreateWindowEv).Error.Error(),
			)})
	})
}

func layoutToolTreeViewEvents(bus *events.Bus, doc *dom.DOM,
	tree *tview.TreeView) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			bus.Emit(ui.ChangeFocusEv{Target: doc.GetElemByID(deviceAPIListID)})
		}

		switch event.Rune() {
		case 'N':
			// Create a new screen
			// I think I will leave this one for later
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

			ref := curNode.GetReference()
			wID, ok := ref.(int16)
			if !ok {
				// Whatever, do something better here
				break
			}

			bus.Emit(ui.DestroyWindowEv{ID: wID})
		case 'm':
			// Go into move mode
		case 'e':
			// Go into edit mode
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

func buildLayoutTreeComponent(bus *events.Bus, doc *dom.DOM) (*dom.UINode, error) {
	layoutTree := tview.NewTreeView()

	layoutTree.SetBorder(true).SetTitle("Layout Tree").
		SetInputCapture(layoutToolTreeViewEvents(bus, doc, layoutTree))

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

func buildLayoutActionPagesComponent(_ *events.Bus, doc *dom.DOM) (*dom.UINode, error) {
	actionPages := tview.NewPages()

	actionPages.SetBorder(true).SetTitle("Action Page")

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

// Layout tool actions

// Validate the form inputs
func validateFormInputs(x, y, w, h, title string) (models.Window, error) {
	xValue, err := strconv.ParseUint(x, 10, 64)
	if err != nil {
		return models.Window{}, err
	}
	yValue, err := strconv.ParseUint(y, 10, 64)
	if err != nil {
		return models.Window{}, err
	}
	widthValue, err := strconv.ParseUint(w, 10, 64)
	if err != nil {
		return models.Window{}, err
	}
	heightValue, err := strconv.ParseUint(h, 10, 64)
	if err != nil {
		return models.Window{}, err
	}

	return models.Window{
		X:      uint16(xValue),
		Y:      uint16(yValue),
		Width:  uint16(widthValue),
		Height: uint16(heightValue),
		Title:  title,
	}, nil
}

func createNewWindowForm(bus *events.Bus, doc *dom.DOM) {
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
			// Validate the inputs
			window, err := validateFormInputs(
				x0.GetText(),
				y0.GetText(),
				width.GetText(),
				height.GetText(),
				title.GetText(),
			)

			if err != nil {
				bus.Emit(ui.LogEv{Log: fmt.Sprintf("failed to parse form: %s\n", err.Error())})
				return
			}

			bus.Emit(ui.CreateWindowEv{Window: window})
		})

	form.SetBorder(true).
		SetTitle("new window form").
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Key() {
			case tcell.KeyEscape:
				bus.Emit(ui.ChangeFocusEv{Target: doc.GetElemByID(layoutToolFlexID)})
			}

			return ev
		})

	var formNode *dom.UINode
	formNode = doc.GetNodeByID("new-window-form")
	if formNode == nil {
		formNode, err = doc.NewUINode("new-window-form",
			doc.GetElemByID(layoutToolActionPagesID), form)
		if err != nil {
			panic("failed to create UI node for the new window form: " + err.Error())
		}
	}

	AddAndShowPage(bus, doc, doc.GetElemByID(layoutToolActionPagesID).(*tview.Pages),
		formNode)
}
