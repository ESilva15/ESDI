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
func validateFormInputs(x, y, w, h,
	title, prev, titleSize, textSize string,
	showID bool) (models.Window, error) {
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
	titleSizeValue, err := strconv.ParseUint(titleSize, 10, 64)
	if err != nil {
		return models.Window{}, err
	}
	textSizeValue, err := strconv.ParseUint(textSize, 10, 64)
	if err != nil {
		return models.Window{}, err
	}

	showIDValue := cdashdisplay.ShowIDFalse
	if showID {
		showIDValue = cdashdisplay.ShowIDTrue
	}

	return models.Window{
		X:            uint16(xValue),
		Y:            uint16(yValue),
		Width:        uint16(widthValue),
		Height:       uint16(heightValue),
		ShowID:       showIDValue,
		TitleSize:    uint8(titleSizeValue),
		TextSize:     uint8(textSizeValue),
		PreviewValue: prev,
		Title:        title,
	}, nil
}

func createNewWindowForm(bus *events.Bus, doc *dom.DOM) {
	var err error

	x0 := tview.NewInputField().SetLabel("x")
	y0 := tview.NewInputField().SetLabel("y")
	width := tview.NewInputField().SetLabel("width")
	height := tview.NewInputField().SetLabel("height")
	title := tview.NewInputField().SetLabel("title")
	previewValue := tview.NewInputField().SetLabel("preview value")
	showID := tview.NewCheckbox().SetLabel("show ID").SetChecked(true)
	winType := tview.NewDropDown().SetLabel("type").
		SetOptions([]string{"string", "bar"}, func(s string, id int) {
			// Here we have to add extra options for the configuration of default
			// values and so on
		}).
		SetCurrentOption(0)
	titleSize := tview.NewDropDown().SetLabel("Title Size").
		SetOptions([]string{"1", "2", "3", "4", "5", "6", "7", "8"}, func(s string, id int) {
		}).
		SetCurrentOption(0)
	textSize := tview.NewDropDown().SetLabel("Text Size").
		SetOptions([]string{"1", "2", "3", "4", "5", "6", "7", "8"}, func(s string, id int) {
		}).
		SetCurrentOption(0)

	form := tview.NewForm().
		AddFormItem(x0).
		AddFormItem(y0).
		AddFormItem(width).
		AddFormItem(height).
		AddFormItem(title).
		AddFormItem(previewValue).
		AddFormItem(showID).
		AddFormItem(winType).
		AddFormItem(titleSize).
		AddFormItem(textSize).
		AddButton("Create", func() {
			_, titleSizeTxt := titleSize.GetCurrentOption()
			_, textSizeTxt := textSize.GetCurrentOption()
			// Validate the inputs
			window, err := validateFormInputs(
				x0.GetText(),
				y0.GetText(),
				width.GetText(),
				height.GetText(),
				title.GetText(),
				previewValue.GetText(),
				titleSizeTxt,
				textSizeTxt,
				showID.IsChecked(),
			)

			if err != nil {
				bus.Emit(ui.LogEv{Log: fmt.Sprintf("failed to parse form: %s\n", err.Error())})
				return
			}

			bus.Emit(ui.LogEv{Log: fmt.Sprintf("showID: %d\n", window.ShowID)})
			bus.Emit(ui.CreateWindowEv{Window: window})
		})

	form.SetBorder(true).
		SetTitle("new window form").
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Key() {
			case tcell.KeyEscape:
				bus.Emit(ui.ChangeFocusEv{Target: doc.GetElemByID(LayoutToolFlexID)})
			}

			return ev
		})

	var formNode *dom.UINode
	formNode = doc.GetNodeByID("new-window-form")
	if formNode == nil {
		formNode, err = doc.NewUINode("new-window-form",
			doc.GetElemByID(LayoutToolActionPagesID), form)
		if err != nil {
			panic("failed to create UI node for the new window form: " + err.Error())
		}
	}

	AddAndShowPage(doc.GetElemByID(LayoutToolActionPagesID).(*tview.Pages), formNode, true)
}

func windowInfoPageID(idx int16) string {
	return fmt.Sprintf("layout-tool-win-info-%d", idx)
}

func windowInfoForm(bus *events.Bus, doc *dom.DOM, idx int16,
	win *cdashdisplay.UIWindow) *dom.UINode {
	var err error

	x0 := tview.NewInputField().SetLabel("x").
		SetText(fmt.Sprintf("%d", win.Dims.X0))
	y0 := tview.NewInputField().SetLabel("y").
		SetText(fmt.Sprintf("%d", win.Dims.Y0))
	width := tview.NewInputField().SetLabel("width").
		SetText(fmt.Sprintf("%d", win.Dims.Width))
	height := tview.NewInputField().SetLabel("height").
		SetText(fmt.Sprintf("%d", win.Dims.Height))
	title := tview.NewInputField().SetLabel("title").
		SetText(win.Title.String())
	previewValue := tview.NewInputField().SetLabel("preview value").
		SetText(win.Opts.PreviewValue.String())
	showID := tview.NewCheckbox().SetLabel("show ID").SetChecked(true)
	winType := tview.NewDropDown().SetLabel("type").
		SetOptions([]string{"string", "bar"}, func(s string, id int) {
			// Here we have to add extra options for the configuration of default
			// values and so on
		}).
		SetCurrentOption(0)
	// NOTE: put this thing into a loop that generate the numbers or something yo
	titleSize := tview.NewDropDown().SetLabel("Title Size").
		SetOptions([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
			"11", "12", "13", "14", "15", "16", "17", "18", "19", "20"},
			func(s string, id int) {
			}).
		SetCurrentOption(int(win.Decor.TitleSize))
	textSize := tview.NewDropDown().SetLabel("Text Size").
		SetOptions([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
			"11", "12", "13", "14", "15", "16", "17", "18", "19", "20"},
			func(s string, id int) {
			}).
		SetCurrentOption(int(win.Decor.TextSize))

	form := tview.NewForm().
		AddFormItem(x0).
		AddFormItem(y0).
		AddFormItem(width).
		AddFormItem(height).
		AddFormItem(title).
		AddFormItem(previewValue).
		AddFormItem(showID).
		AddFormItem(titleSize).
		AddFormItem(textSize).
		AddFormItem(winType).
		AddButton("Update", func() {
			bus.Emit(ui.LogEv{Log: "update event was created\n"})

			_, titleSizeTxt := titleSize.GetCurrentOption()
			_, textSizeTxt := textSize.GetCurrentOption()
			// Validate the inputs
			window, err := validateFormInputs(
				x0.GetText(),
				y0.GetText(),
				width.GetText(),
				height.GetText(),
				title.GetText(),
				previewValue.GetText(),
				titleSizeTxt,
				textSizeTxt,
				showID.IsChecked(),
			)

			if err != nil {
				bus.Emit(ui.LogEv{Log: fmt.Sprintf("failed to parse form: %s\n", err.Error())})
				return
			}

			bus.Emit(ui.LogEv{Log: fmt.Sprintf("showID: %d\n", window.ShowID)})
			bus.Emit(ui.LogEv{Log: "sending update window ev\n"})
			bus.Emit(ui.UpdateWindowEv{ID: idx, Window: window})
		})

	form.SetBorder(true).
		SetTitle(fmt.Sprintf("Info - %s [%2d]", win.Title.String(), idx)).
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Key() {
			case tcell.KeyEscape:
				bus.Emit(ui.ChangeFocusEv{Target: doc.GetElemByID(LayoutToolTreeID)})
			}

			return ev
		})

	// In case the window is moved we need to update this form
	bus.On(ui.WindowMovedEv{}, func(e any) {
		data := e.(ui.WindowMovedEv)
		// Only update the fields if the event is for this thing
		if data.ID != idx {
			return
		}

		// Get the current window-info-page
		elem := doc.GetElemByID(windowInfoPageID(data.ID))
		if elem == nil {
			bus.Emit(ui.LogEv{Log: "could't get hold of info page"})
			return
		}

		x0.SetText(fmt.Sprintf("%d", data.Dims.X0))
		x0.SetText(fmt.Sprintf("%d", data.Dims.Y0))
		width.SetText(fmt.Sprintf("%d", data.Dims.Width))
		height.SetText(fmt.Sprintf("%d", data.Dims.Height))
	})

	var formNode *dom.UINode
	elemID := windowInfoPageID(idx)
	formNode = doc.GetNodeByID(elemID)
	layoutToolActionPagesElem := doc.GetElemByID(LayoutToolActionPagesID)
	if formNode == nil {
		formNode, err = doc.NewUINode(elemID, layoutToolActionPagesElem, form)
		if err != nil {
			panic("failed to create UI node for the new window form: " + err.Error())
		}
	}

	return formNode
}
