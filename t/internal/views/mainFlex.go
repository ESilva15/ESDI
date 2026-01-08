// Package views
package views

import (
	"esdi/t/internal/dom"
	"fmt"

	"github.com/rivo/tview"
)

func buildRightSidePages(root *dom.DOM, ctx *UIContext) (*dom.UINode, error) {
	// We need to build a set of pages with an empty page
	emptyPage := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			ctx.Redraw()
		})
	emptyPage.SetBorder(true).SetTitle("-- Tool Area --")

	fmt.Fprintf(emptyPage, "No Tool Selected")

	apiToolPages := tview.NewPages().
		AddPage("emptypage", emptyPage, true, true)

	apiToolPagesNode, err := root.NewUINode("api-tool-pages", root.GetElemByID("right-flex"),
		apiToolPages)
	if err != nil {
		return nil, err
	}

	return apiToolPagesNode, nil
}

func buildRightSideFlex(root *dom.DOM, ctx *UIContext) (*dom.UINode, error) {
	apiToolPagesNode, err := buildRightSidePages(root, ctx)
	if err != nil {
		return nil, err
	}

	// This will be the right side flex
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flexNode, err := root.NewUINode("right-flex", root.GetElemByID("main-flex"), flex)
	if err != nil {
		return nil, err
	}

	// Make the output window
	outputWin := tview.NewTextView().SetBorder(true).SetTitle("DebugWindow")
	outputWinNode, err := root.NewUINode("debug-window", root.GetRootElem(),
		outputWin)
	if err != nil {
		return nil, err
	}

	flex.
		AddItem(apiToolPagesNode.Self, 0, 5, false).
		AddItem(outputWinNode.Self, 0, 2, false)

	return flexNode, nil
}

func BuildMainFlex(root *dom.DOM, ctx *UIContext) (*tview.Flex, error) {
	// To build the main view we must set the DOM root
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	mainFlexUINode, err := root.NewUINode("main-flex", nil, mainFlex)
	if err != nil {
		return nil, err
	}

	root.SetRoot(mainFlexUINode)

	deviceAPIList := tview.NewList().
		AddItem("layout", "build a layout for CDashDisplay", 0, func() {
			// When clicked we want to replace the right side pages with this
		})
	deviceAPIList.SetBorder(true).SetTitle("list")

	apiListWindowUINode, err := root.NewUINode("device-api-list", root.GetRootElem(),
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
