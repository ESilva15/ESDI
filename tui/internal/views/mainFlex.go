// Package views
package views

import (
	"fmt"

	"esdi/tui/internal/controllers"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/ui"

	"github.com/rivo/tview"
)

const (
	mainFlexID      = "main-flex"
	apiToolPagesID  = "api-tool-pages"
	rightFlexID     = "right-flex"
	outputPaneID    = "output-window"
	deviceAPIListID = "device-api-list"
)

const (
	pageName = "empty-page"
)

func buildRightSidePages(root *dom.DOM, ctx *ui.UIContext) (*dom.UINode, error) {
	// We need to build a set of pages with an empty page
	emptyPage := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			ctx.Redraw()
		})
	emptyPage.SetBorder(true).SetTitle("-- Tool Area --")

	fmt.Fprintf(emptyPage, "No Tool Selected")

	apiToolPages := tview.NewPages().
		AddPage(pageName, emptyPage, true, true)

	apiToolPagesNode, err := root.NewUINode(apiToolPagesID, root.GetElemByID(rightFlexID),
		apiToolPages)
	if err != nil {
		return nil, err
	}

	return apiToolPagesNode, nil
}

func buildRightSideFlex(root *dom.DOM, ctx *ui.UIContext) (*dom.UINode, error) {
	apiToolPagesNode, err := buildRightSidePages(root, ctx)
	if err != nil {
		return nil, err
	}

	// This will be the right side flex
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flexNode, err := root.NewUINode(rightFlexID, root.GetElemByID(mainFlexID), flex)
	if err != nil {
		return nil, err
	}

	outputWin := root.GetElemByID(outputPaneID)
	if outputWin == nil {
		panic("failed to attach output win to UI")
	}

	flex.
		AddItem(apiToolPagesNode.Self, 0, 5, false).
		AddItem(outputWin, 0, 2, false)

	return flexNode, nil
}

func BuildMainFlex(root *dom.DOM, ctx *ui.UIContext,
	wc *controllers.WindowingController) (*tview.Flex, error) {
	// To build the main view we must set the DOM root
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	mainFlexUINode, err := root.NewUINode(mainFlexID, nil, mainFlex)
	if err != nil {
		return nil, err
	}

	root.SetRoot(mainFlexUINode)

	deviceAPIList := tview.NewList().
		AddItem("layout", "build a layout for CDashDisplay", 0, func() {
			layoutToolUIOnSelect(root, ctx, wc)
		})
	deviceAPIList.SetBorder(true).SetTitle("list")

	apiListWindowUINode, err := root.NewUINode(deviceAPIListID, root.GetRootElem(),
		deviceAPIList)
	if err != nil {
		return nil, err
	}

	rightSideFlex, err := buildRightSideFlex(root, ctx)
	if err != nil {
		return nil, err
	}

	mainFlex.
		AddItem(apiListWindowUINode.Self, 0, 1, false).
		AddItem(rightSideFlex.Self, 0, 4, false)

	return mainFlex, nil
}
