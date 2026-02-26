package controllers

import (
	"esdi/cdashdisplay"
	"esdi/tui/internal/views"

	"github.com/gdamore/tcell/v2"
)

type LayoutController struct {
	*Controller
	OnExit         func()
	LayoutToolView *views.LayoutToolView
	Messages       chan string
}

func NewLayoutController(base *Controller) *LayoutController {
	lc := &LayoutController{
		Controller:     base,
		LayoutToolView: views.NewLayoutToolView(),
		Messages:       make(chan string, 10),
	}

	lc.registerHooks()

	return lc
}

func (lc *LayoutController) registerHooks() {
	lc.LayoutToolView.LayoutTree.Tree.SetInputCapture(
		func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Key() {
			case tcell.KeyEsc:
				lc.OnExit()
			}

			switch ev.Rune() {
			case 'n':
				// Create a new window
				lc.Messages <- "calling new window form\n"
				lc.newWindowAction()
			case 'x':
				lc.Messages <- "calling delete window\n"
				// Delete selected window
				// bus.Emit(ui.LogEv{Log: })
				// wRef, err := getCurNodeRef(tree)
				// if err != nil {
				// 	bus.Emit(ui.LogEv{Log: " -> " + err.Error()})
				// 	break
				// }
				// bus.Emit(ui.DestroyWindowEv{ID: wRef.ID})
			case 'm':
				lc.Messages <- "calling window manipulation tool\n"
				// Go into move mode
				// bus.Emit(ui.LogEv{Log: })
				// wRef, err := getCurNodeRef(tree)
				// if err != nil {
				// 	bus.Emit(ui.LogEv{Log: " -> " + err.Error()})
				// 	break
				// }
				// windowManipulationTool(bus, doc, wRef.ID)
			case 'e':
				// Go into edit mode
			case 's':
				// Save the current layout
				// bus.Emit(ui.SaveLayoutEv{})
			case 'l':
				// Load the layout
				// bus.Emit(ui.LoadLayoutEv{})
			case 'g':
				// Go -> launches the current set up source
				// streamingWindow(bus, doc)
			}

			return ev
		},
	)
}

func (lc *LayoutController) parseCreateWindowInput() (*cdashdisplay.UIWindow, error) {
	form := lc.LayoutToolView.LayoutActions.CreateWindowView

	lc.Messages <- form.Form.X.GetText()

	// xValue, err := strconv.ParseUint(x, 10, 64)
	// if err != nil {
	// 	return models.Window{}, err
	// }
	// yValue, err := strconv.ParseUint(y, 10, 64)
	// if err != nil {
	// 	return models.Window{}, err
	// }
	// widthValue, err := strconv.ParseUint(w, 10, 64)
	// if err != nil {
	// 	return models.Window{}, err
	// }
	// heightValue, err := strconv.ParseUint(h, 10, 64)
	// if err != nil {
	// 	return models.Window{}, err
	// }
	// titleSizeValue, err := strconv.ParseUint(titleSize, 10, 64)
	// if err != nil {
	// 	return models.Window{}, err
	// }
	// textSizeValue, err := strconv.ParseUint(textSize, 10, 64)
	// if err != nil {
	// 	return models.Window{}, err
	// }
	//
	// showIDValue := cdashdisplay.ShowIDFalse
	// if showID {
	// 	showIDValue = cdashdisplay.ShowIDTrue
	// }

	// 	win := e.(ui.CreateWindowEv).Window
	//
	// 	winDecor := cdashdisplay.DefaultDecorations
	// 	winDecor.TextSize = win.TextSize
	// 	winDecor.TitleSize = win.TitleSize
	//
	// 	uiWindow := cdashdisplay.UIWindow{
	// 		Dims: cdashdisplay.UIDimensions{
	// 			X0:     win.X,
	// 			Y0:     win.Y,
	// 			Width:  win.Width,
	// 			Height: win.Height,
	// 		},
	// 		Opts: cdashdisplay.UIWindowOpts{
	// 			WinType:      win.Type, // NOTE: values aren't implemented yet
	// 			ShowID:       win.ShowID,
	// 			PreviewValue: helper.B32(win.PreviewValue),
	// 		},
	// 		Decor: winDecor,
	// 		Title: helper.B32(win.Title),
	// 	}
	//
	// 	wID, err := mc.CDash.CreateWindow(uiWindow)
	// 	if err != nil {
	// 		mc.Bus.Emit(ui.PrintLogEv{Log: "failed to create window\n"})
	// 		return
	// 	}
	//
	// 	mc.Bus.Emit(ui.WindowCreatedEv{ID: wID, Win: uiWindow})
	// 	mc.Bus.Emit(ui.PrintLogEv{Log: "Window created!\n"})

	// return models.Window{
	// 	X:            uint16(xValue),
	// 	Y:            uint16(yValue),
	// 	Width:        uint16(widthValue),
	// 	Height:       uint16(heightValue),
	// 	ShowID:       showIDValue,
	// 	TitleSize:    uint8(titleSizeValue),
	// 	TextSize:     uint8(textSizeValue),
	// 	PreviewValue: prev,
	// 	Title:        title,
	// }, nil

	return nil, nil
}

func (lc *LayoutController) newWindowAction() {
	if lc.LayoutToolView.LayoutActions.CreateWindowView != nil {
		// already exists, just change focus and leave
		lc.App.SetFocus(lc.LayoutToolView.LayoutActions.CreateWindowView.Form.Form)
		lc.Messages <- "already exists"
		return
	}

	// Get the new window form view
	newWindowForm := views.NewCreateWindowFormView()
	lc.LayoutToolView.LayoutActions.CreateWindowView = newWindowForm

	// Set the event capture for this form
	newWindowForm.Form.Form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEscape:
			lc.App.SetFocus(lc.LayoutToolView.LayoutTree.Tree)
		}

		return ev
	})

	// Set the behaviour for when we press the create button
	newWindowForm.CreateBtn.SetSelectedFunc(func() {
		lc.Messages <- "am I here???"
		lc.parseCreateWindowInput()
		// _, titleSizeTxt := titleSize.GetCurrentOption()
		// _, textSizeTxt := textSize.GetCurrentOption()
		// // Validate the inputs
		// window, err := validateFormInputs(
		// 	x0.GetText(),
		// 	y0.GetText(),
		// 	width.GetText(),
		// 	height.GetText(),
		// 	title.GetText(),
		// 	previewValue.GetText(),
		// 	titleSizeTxt,
		// 	textSizeTxt,
		// 	showID.IsChecked(),
		// )

		// if err != nil {
		// 	// bus.Emit(ui.LogEv{Log: fmt.Sprintf("failed to parse form: %s\n", err.Error())})
		// 	return
		// }

		// bus.Emit(ui.LogEv{Log: fmt.Sprintf("showID: %d\n", window.ShowID)})
		// bus.Emit(ui.CreateWindowEv{Window: window})
	})

	// Set this on the action page and change to it
	views.AddAndShowPage(
		lc.LayoutToolView.LayoutActions.Pages,
		"new-window-action-form",
		newWindowForm.Form.Form,
	)

	// Set its focus
	lc.App.SetFocus(newWindowForm.Form.Form)
}
