// Package views
package views

import (
	"fmt"

	"esdi/tui/internal/dom"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	MainFlexID      = "main-flex"
	APIToolPagesID  = "api-tool-pages"
	RightFlexID     = "right-flex"
	OutputPaneID    = "output-window"
	DeviceAPIListID = "device-api-list"
)

const (
	EmptyPageName = "empty-page"
)

type DeviceAPIToolView struct {
	Pages     *tview.Pages
	ChangedFn func()
}

func NewDeviceAPIToolView() *DeviceAPIToolView {
	// We need to build a set of pages with an empty page
	emptyPage := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			// bus.Emit(ui.RedrawEv{})
		})
	emptyPage.SetBorder(true).SetTitle("-- Tool Area --")
	fmt.Fprintf(emptyPage, "No Tool Selected")

	apiToolPages := tview.NewPages().AddPage(EmptyPageName, emptyPage, true, true)

	return &DeviceAPIToolView{
		Pages:     apiToolPages,
		ChangedFn: func() {}, // blank function to be hooked on
	}
}

type DeviceAPIListView struct {
	List         *tview.List
	InputCapture func(*tcell.EventKey) *tcell.EventKey
	ItemOnSelect func()
}

func NewDeviceAPIListView() *DeviceAPIListView {
	deviceAPIList := tview.NewList().
		AddItem("layout", "build a layout for CDashDisplay", 0, func() {
			// layoutToolUIOnSelect(bus, doc)
		})
	deviceAPIList.SetBorder(true).SetTitle("list")

	return &DeviceAPIListView{
		List:         deviceAPIList,
		InputCapture: func(ev *tcell.EventKey) *tcell.EventKey { return ev },
		ItemOnSelect: func() {}, // blank functions, they have to be hooked on by the controller
	}
}

type DeviceAPIView struct {
	MainFlex       *tview.Flex
	DevAPIList     *DeviceAPIListView
	DevAPIToolView *DeviceAPIToolView
	OutputWindow   *OutputWinView
}

func (dl *DeviceAPIListView) AddItem(name, description string, onSelect func()) {
	dl.List.AddItem(name, description, 0, onSelect)
}

func NewDeviceAPIView(doc *dom.DOM) (*DeviceAPIView, error) {
	// To build the main view we must set the DOM root
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	mainFlexUINode, err := doc.NewUINode(MainFlexID, nil, mainFlex)
	if err != nil {
		return nil, err
	}

	doc.SetRoot(mainFlexUINode)

	deviceAPIList := NewDeviceAPIListView()
	apiListWindowUINode, err := doc.NewUINode(DeviceAPIListID, doc.GetRootElem(),
		deviceAPIList.List)
	if err != nil {
		return nil, err
	}

	apiToolPages := NewDeviceAPIToolView()
	apiToolPagesNode, err := doc.NewUINode(APIToolPagesID, doc.GetElemByID(RightFlexID),
		apiToolPages.Pages)
	if err != nil {
		return nil, err
	}

	mainFlex.
		AddItem(apiListWindowUINode.Self, 0, 1, false).
		AddItem(apiToolPagesNode.Self, 0, 4, false)

	// Output window

	outputWin := NewOutputWinView()
	_, err = doc.NewUINode("output-window", nil, outputWin.TextArea)
	if err != nil {
		return nil, err
	}

	// Flex with debug window
	// --------------------------------------------------------------------------
	flexWithOutputWindow := tview.NewFlex().SetDirection(tview.FlexRow)
	flexWithOutputWindow.
		AddItem(mainFlex, 0, 5, true).
		AddItem(doc.GetElemByID(OutputPaneID), 0, 2, false)
	// --------------------------------------------------------------------------

	return &DeviceAPIView{
		MainFlex:       flexWithOutputWindow,
		DevAPIList:     deviceAPIList,
		DevAPIToolView: apiToolPages,
		OutputWindow:   outputWin,
	}, nil
}
