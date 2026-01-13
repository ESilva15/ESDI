// Package views
package views

import (
	"fmt"

	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
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

func buildRightSidePages(bus *events.Bus, doc *dom.DOM) (*dom.UINode, error) {
	// We need to build a set of pages with an empty page
	emptyPage := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			bus.Emit(ui.RedrawEv{})
		})
	emptyPage.SetBorder(true).SetTitle("-- Tool Area --")

	fmt.Fprintf(emptyPage, "No Tool Selected")

	apiToolPages := tview.NewPages().
		AddPage(pageName, emptyPage, true, true)

	apiToolPagesNode, err := doc.NewUINode(apiToolPagesID, doc.GetElemByID(rightFlexID),
		apiToolPages)
	if err != nil {
		return nil, err
	}

	return apiToolPagesNode, nil
}

func buildRightSideFlex(bus *events.Bus, doc *dom.DOM) (*dom.UINode, error) {
	apiToolPagesNode, err := buildRightSidePages(bus, doc)
	if err != nil {
		return nil, err
	}

	// This will be the right side flex
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flexNode, err := doc.NewUINode(rightFlexID, doc.GetElemByID(mainFlexID), flex)
	if err != nil {
		return nil, err
	}

	outputWin := doc.GetElemByID(outputPaneID)
	if outputWin == nil {
		panic("failed to attach output win to UI")
	}

	flex.
		AddItem(apiToolPagesNode.Self, 0, 5, false).
		AddItem(outputWin, 0, 2, false)

	return flexNode, nil
}

func BuildMainFlex(bus *events.Bus, doc *dom.DOM) (*tview.Flex, error) {
	// To build the main view we must set the DOM root
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	mainFlexUINode, err := doc.NewUINode(mainFlexID, nil, mainFlex)
	if err != nil {
		return nil, err
	}

	doc.SetRoot(mainFlexUINode)

	deviceAPIList := tview.NewList().
		AddItem("layout", "build a layout for CDashDisplay", 0, func() {
			layoutToolUIOnSelect(bus, doc)
		})
	deviceAPIList.SetBorder(true).SetTitle("list")

	apiListWindowUINode, err := doc.NewUINode(deviceAPIListID, doc.GetRootElem(),
		deviceAPIList)
	if err != nil {
		return nil, err
	}

	rightSideFlex, err := buildRightSideFlex(bus, doc)
	if err != nil {
		return nil, err
	}

	mainFlex.
		AddItem(apiListWindowUINode.Self, 0, 1, false).
		AddItem(rightSideFlex.Self, 0, 4, false)

	return mainFlex, nil
}
