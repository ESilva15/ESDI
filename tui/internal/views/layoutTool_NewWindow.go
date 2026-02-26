package views

import (
	"fmt"

	"esdi/cdashdisplay"

	"github.com/rivo/tview"
)

type CDashDisplayWindowFormView struct {
	Form         *tview.Form
	X            *tview.InputField
	Y            *tview.InputField
	Width        *tview.InputField
	Height       *tview.InputField
	Title        *tview.InputField
	PreviewValue *tview.InputField
	ShowID       *tview.Checkbox
	WinType      *tview.DropDown
	TitleSize    *tview.DropDown
	TextSize     *tview.DropDown
}

func NewCDashDisplayWindowFormView() *CDashDisplayWindowFormView {
	view := &CDashDisplayWindowFormView{}

	view.X = tview.NewInputField().SetLabel("x")
	view.Y = tview.NewInputField().SetLabel("y")
	view.Width = tview.NewInputField().SetLabel("width")
	view.Height = tview.NewInputField().SetLabel("height")
	view.Title = tview.NewInputField().SetLabel("title")
	view.PreviewValue = tview.NewInputField().SetLabel("preview value")
	view.ShowID = tview.NewCheckbox().SetLabel("show ID").SetChecked(true)
	view.WinType = tview.NewDropDown().SetLabel("type").
		SetOptions([]string{"string", "bar"}, func(s string, id int) {}).
		SetCurrentOption(0)
	view.TitleSize = tview.NewDropDown().SetLabel("Title Size").SetCurrentOption(0)
	for k := range 20 {
		view.TitleSize.AddOption(fmt.Sprintf("%d", k+1), blankDropdownOptionCallback)
	}
	view.TitleSize.SetCurrentOption(0)

	view.TextSize = tview.NewDropDown().SetLabel("Text Size")
	for k := range 20 {
		view.TextSize.AddOption(fmt.Sprintf("%d", k+1), blankDropdownOptionCallback)
	}
	view.TitleSize.SetCurrentOption(0)

	view.Form = tview.NewForm().
		AddFormItem(view.X).
		AddFormItem(view.Y).
		AddFormItem(view.Width).
		AddFormItem(view.Height).
		AddFormItem(view.Title).
		AddFormItem(view.PreviewValue).
		AddFormItem(view.ShowID).
		AddFormItem(view.WinType).
		AddFormItem(view.TitleSize).
		AddFormItem(view.TextSize)

	view.Form.SetBorder(true).
		SetTitle("New Window Form").
		SetTitleAlign(tview.AlignLeft)
		// Need to inject the event capture later

	return view
}

type CreateWindowFormView struct {
	Form      *CDashDisplayWindowFormView
	CreateBtn *tview.Button
}

func NewCreateWindowFormView() *CreateWindowFormView {
	view := &CreateWindowFormView{
		Form: NewCDashDisplayWindowFormView(),
	}

	view.CreateBtn = tview.NewButton("Create")
	// Need to inject the button functionality later

	view.Form.Form.AddButton("Create", func() {})

	return view
}

type WindowFormView struct {
	Form      *CDashDisplayWindowFormView
	UpdateBtn *tview.Button
}

func NewWindowFormView(win *cdashdisplay.UIWindow) *WindowFormView {
	view := &WindowFormView{
		Form: NewCDashDisplayWindowFormView(),
	}

	view.UpdateBtn = tview.NewButton("Update")
	// Need to inject the button functionality later
	// AddButton("Update", func() {
	// 	bus.Emit(ui.LogEv{Log: "update event was created\n"})
	//
	// 	_, titleSizeTxt := titleSize.GetCurrentOption()
	// 	_, textSizeTxt := textSize.GetCurrentOption()
	// 	// Validate the inputs
	// 	window, err := validateFormInputs(
	// 		x0.GetText(),
	// 		y0.GetText(),
	// 		width.GetText(),
	// 		height.GetText(),
	// 		title.GetText(),
	// 		previewValue.GetText(),
	// 		titleSizeTxt,
	// 		textSizeTxt,
	// 		showID.IsChecked(),
	// 	)
	//
	// 	if err != nil {
	// 		bus.Emit(ui.LogEv{Log: fmt.Sprintf("failed to parse form: %s\n", err.Error())})
	// 		return
	// 	}
	//
	// 	bus.Emit(ui.LogEv{Log: fmt.Sprintf("showID: %d\n", window.ShowID)})
	// 	bus.Emit(ui.LogEv{Log: "sending update window ev\n"})
	// 	bus.Emit(ui.UpdateWindowEv{ID: idx, Window: window})
	// })

	// In case the window is moved we need to update this form
	// bus.On(ui.WindowMovedEv{}, func(e any) {
	// 	data := e.(ui.WindowMovedEv)
	// 	// Only update the fields if the event is for this thing
	// 	if data.ID != idx {
	// 		return
	// 	}
	//
	// 	// Get the current window-info-page
	// 	elem := doc.GetElemByID(windowInfoPageID(data.ID))
	// 	if elem == nil {
	// 		bus.Emit(ui.LogEv{Log: "could't get hold of info page"})
	// 		return
	// 	}
	//
	// 	x0.SetText(fmt.Sprintf("%d", data.Dims.X0))
	// 	x0.SetText(fmt.Sprintf("%d", data.Dims.Y0))
	// 	width.SetText(fmt.Sprintf("%d", data.Dims.Width))
	// 	height.SetText(fmt.Sprintf("%d", data.Dims.Height))
	// })

	// Set the fields data
	view.Form.X.SetText(fmt.Sprintf("%d", win.Dims.X0))
	view.Form.Y.SetText(fmt.Sprintf("%d", win.Dims.Y0))
	view.Form.Width.SetText(fmt.Sprintf("%d", win.Dims.Width))
	view.Form.Height.SetText(fmt.Sprintf("%d", win.Dims.Height))
	view.Form.Title.SetText(win.Title.String())
	view.Form.PreviewValue.SetText(win.Opts.PreviewValue.String())
	view.Form.ShowID.SetChecked(win.Opts.ShowID == 1)
	view.Form.WinType.SetCurrentOption(0) // NOTE: this needs to set the correct option
	view.Form.TitleSize.SetCurrentOption(int(win.Decor.TitleSize))
	view.Form.TextSize.SetCurrentOption(int(win.Decor.TextSize))

	return view
}

func WindowInfoPageID(idx int16) string {
	return fmt.Sprintf("layout-tool-win-info-%d", idx)
}
