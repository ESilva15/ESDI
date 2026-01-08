// Package tui
package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type layout struct {
	root *tview.Flex
}

type devicePane struct {
	apiPane    *tview.List
	actionPane *tview.Pages
	flexPane   *tview.Flex
}

func NewDevicePane(t *tui) *devicePane {
	dp := &devicePane{}

	dp.actionPane = tview.NewPages()

	dp.apiPane = tview.NewList().
		AddItem("destroy-window", "", 0, func() {
			dp.actionPane.RemovePage("action")
			dp.actionPane.AddPage("action",
				tview.NewBox().SetBorder(true).SetTitle("destroy"),
				true, true,
			)
		}).
		AddItem("layout", "", 0, func() {
			createLayoutUI(t, dp)
		})

	dp.apiPane.SetBorder(true).SetTitle("API")

	dp.flexPane = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(dp.apiPane, 0, 1, true).
		AddItem(dp.actionPane, 0, 4, false)

	return dp
}

type debugPane struct {
	view *tview.TextView
}

func NewDebugPane(t *tui) *debugPane {
	output := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			t.debug.view.ScrollToEnd()
			t.app.Draw()
		})
	output.SetBorder(true).SetTitle("Output:")

	return &debugPane{
		view: output,
	}
}

type tui struct {
	app    *tview.Application
	layout *layout
	debug  *debugPane
	device *devicePane
}

func (t *tui) log(formatter string, args ...any) {
	fmt.Fprintf(t.debug.view, formatter, args...)
}

func NewTUI() *tui {
	t := &tui{}

	debug := NewDebugPane(t)
	device := NewDevicePane(t)

	// tui.registerDevices()
	//
	// tui.esdi.PeripheralClerk.FindDevices()
	// tui.populateDeviceList()

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(device.flexPane, 0, 2, true).
		AddItem(debug.view, 0, 1, false)

	t.app = tview.NewApplication()
	t.layout = &layout{
		root: mainFlex,
	}
	t.debug = debug
	t.device = device

	return t
}

func (t *tui) Start() error {
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
			t.app.Stop()
			return nil
		}

		return event
	})

	err := t.app.SetRoot(t.layout.root, true).
		SetFocus(t.device.apiPane).Run()
	if err != nil {
		return err
	}

	return nil
}

func Run() error {
	// Prepare the UI here
	newTui := NewTUI()
	return newTui.Start()
}
