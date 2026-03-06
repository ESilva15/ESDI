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
	view.TextSize.SetCurrentOption(0)

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

	// Set the fields data
	view.SetValues(win)

	view.Form.Form.AddButton("Update", func() {})

	return view
}

func (fv *WindowFormView) SetValues(win *cdashdisplay.UIWindow) {
	fv.Form.X.SetText(fmt.Sprintf("%d", win.Dims.X0))
	fv.Form.Y.SetText(fmt.Sprintf("%d", win.Dims.Y0))
	fv.Form.Width.SetText(fmt.Sprintf("%d", win.Dims.Width))
	fv.Form.Height.SetText(fmt.Sprintf("%d", win.Dims.Height))
	fv.Form.Title.SetText(win.Title.String())
	fv.Form.PreviewValue.SetText(win.Opts.PreviewValue.String())
	fv.Form.ShowID.SetChecked(win.Opts.ShowID == 1)
	fv.Form.WinType.SetCurrentOption(0) // NOTE: this needs to set the correct option
	fv.Form.TitleSize.SetCurrentOption(int(win.Decor.TitleSize))
	fv.Form.TextSize.SetCurrentOption(int(win.Decor.TextSize))
}

func windowInfoPageID(idx int16) string {
	return fmt.Sprintf("layout-tool-win-info-%d", idx)
}

func windowInfoPageTitle(idx int16, title string) string {
	return fmt.Sprintf("%s [%02d]", title, idx)
}
