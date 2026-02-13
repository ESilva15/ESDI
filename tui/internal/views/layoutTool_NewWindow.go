package views

import (
	"fmt"
	"strconv"

	"esdi/cdashdisplay"
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/models"
	"esdi/tui/internal/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
	winType := tview.NewDropDown().SetLabel("type").
		SetOptions([]string{"string", "bar"}, func(s string, id int) {
			// Here we have to add extra options for the configuration of default
			// values and so on
		}).
		SetCurrentOption(0)

	form := tview.NewForm().
		AddFormItem(x0).
		AddFormItem(y0).
		AddFormItem(width).
		AddFormItem(height).
		AddFormItem(title).
		AddFormItem(winType).
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
		formNode, true)
}

func windowInfoPageID(idx int16) string {
	return fmt.Sprintf("layout-tool-win-info-%d", idx)
}

func windowInfoForm(bus *events.Bus, doc *dom.DOM, idx int16,
	win *cdashdisplay.UIWindow) *dom.UINode {
	var err error

	x0 := tview.NewInputField().SetLabel("x").SetText(fmt.Sprintf("%d", win.Dims.X0))
	y0 := tview.NewInputField().SetLabel("y").SetText(fmt.Sprintf("%d", win.Dims.Y0))
	width := tview.NewInputField().SetLabel("width").SetText(fmt.Sprintf("%d", win.Dims.Width))
	height := tview.NewInputField().SetLabel("height").SetText(fmt.Sprintf("%d", win.Dims.Height))
	title := tview.NewInputField().SetLabel("title").SetText(win.Title.String())
	winType := tview.NewDropDown().SetLabel("type").
		SetOptions([]string{"string", "bar"}, func(s string, id int) {
			// Here we have to add extra options for the configuration of default
			// values and so on
		}).
		SetCurrentOption(0)

	form := tview.NewForm().
		AddFormItem(x0).
		AddFormItem(y0).
		AddFormItem(width).
		AddFormItem(height).
		AddFormItem(title).
		AddFormItem(winType).
		AddButton("Update", func() {
			bus.Emit(ui.LogEv{Log: "update event was created\n"})
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

			bus.Emit(ui.LogEv{Log: "sending update window ev\n"})
			bus.Emit(ui.UpdateWindowEv{ID: idx, Window: window})
		})

	form.SetBorder(true).
		SetTitle(fmt.Sprintf("Info - %s [%2d]", win.Title.String(), idx)).
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Key() {
			case tcell.KeyEscape:
				bus.Emit(ui.ChangeFocusEv{Target: doc.GetElemByID(layoutToolTreeID)})
			}

			return ev
		})

	var formNode *dom.UINode
	elemId := windowInfoPageID(idx)
	formNode = doc.GetNodeByID(elemId)
	layoutToolActionPagesElem := doc.GetElemByID(layoutToolActionPagesID)
	if formNode == nil {
		formNode, err = doc.NewUINode(elemId, layoutToolActionPagesElem, form)
		if err != nil {
			panic("failed to create UI node for the new window form: " + err.Error())
		}
	}

	// AddAndShowPage(bus, doc, layoutToolActionPagesElem.(*tview.Pages), formNode)
	return formNode
}
