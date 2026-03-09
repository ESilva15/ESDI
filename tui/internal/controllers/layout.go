package controllers

import (
	"esdi/cdashdisplay"
	helper "esdi/helpers"
	"esdi/telemetry"
	"esdi/tui/internal/services"
	"esdi/tui/internal/views"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LayoutController struct {
	*Controller
	OnExit         func()
	LayoutToolView *views.LayoutToolView
	Messages       chan string
	DevService     *services.CDashService
	MoveToolState  *windowManipState
}

func NewLayoutController(base *Controller, service *services.CDashService) *LayoutController {
	lc := &LayoutController{
		Controller:     base,
		LayoutToolView: views.NewLayoutToolView(),
		Messages:       make(chan string, 10),
		DevService:     service,
		MoveToolState:  &windowManipState{Mode: moveMode},
	}

	lc.registerHooks()

	return lc
}

func (lc *LayoutController) registerHooks() {
	// Set the input capture behaviour
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
				lc.deleteWindow()
			case 'm':
				lc.Messages <- "calling window manipulation tool\n"
				lc.moveWindow()
			case 's':
				// Save the current layout
				lc.Messages <- "calling save layout\n"
				lc.saveLayout()
			case 'l':
				// Load the layout
				lc.Messages <- "calling load layout\n"
				lc.loadLayout()
			case 'g':
				// Go -> launches the current set up source
				// streamingWindow(bus, doc)
			}

			return ev
		},
	)

	// Set the tree input actions
	lc.LayoutToolView.LayoutTree.Tree.SetChangedFunc(func(node *tview.TreeNode) {
		// Here we want to change the current existing form on the action pages
		nodeRef := node.GetReference()
		if nodeRef == nil {
			// its probably root I guess, thats what I'll believe
			return
		}

		nodeWinRef := nodeRef.(int16)
		lc.Messages <- fmt.Sprintf("changing to -> %d\n", nodeWinRef)

		lc.LayoutToolView.ShowWindowFormByID(nodeWinRef)
	})

	lc.LayoutToolView.LayoutTree.Tree.SetSelectedFunc(func(node *tview.TreeNode) {
		// Change focus to the selected window
		nodeRef := node.GetReference()
		if nodeRef == nil {
			// its probably root I guess, thats what I'll believe
			return
		}

		nodeWinRef := nodeRef.(int16)

		lc.App.SetFocus(lc.LayoutToolView.FormQuickAccess[nodeWinRef].Form.Form)
	})
}

func (lc *LayoutController) parseWindowFormData(
	form views.CDashDisplayWindowFormView,
) (*cdashdisplay.DesktopUIWindow, error) {

	xValue, err := strconv.ParseUint(form.X.GetText(), 10, 64)
	if err != nil {
		return nil, err
	}
	yValue, err := strconv.ParseUint(form.Y.GetText(), 10, 64)
	if err != nil {
		return nil, err
	}
	widthValue, err := strconv.ParseUint(form.Width.GetText(), 10, 64)
	if err != nil {
		return nil, err
	}
	heightValue, err := strconv.ParseUint(form.Height.GetText(), 10, 64)
	if err != nil {
		return nil, err
	}

	titleSizeInputID, titleSizeInput := form.TitleSize.GetCurrentOption()
	if titleSizeInputID == -1 {
		return nil, fmt.Errorf("no option selected for title size")
	}
	titleSizeValue, err := strconv.ParseUint(titleSizeInput, 10, 64)
	if err != nil {
		return nil, err
	}

	textSizeInputID, textSizeInput := form.TextSize.GetCurrentOption()
	if textSizeInputID == -1 {
		return nil, fmt.Errorf("no option selected for text size")
	}
	textSizeValue, err := strconv.ParseUint(textSizeInput, 10, 64)
	if err != nil {
		return nil, err
	}

	telemFieldInputID, telemField := form.TelemetryField.GetCurrentOption()
	if telemFieldInputID == -1 {
		return nil, fmt.Errorf("no option selected for telemetry field")
	}

	showIDValue := cdashdisplay.ShowIDFalse
	if form.ShowID.IsChecked() {
		showIDValue = cdashdisplay.ShowIDTrue
	}

	winDecor := cdashdisplay.DefaultDecorations
	winDecor.TextSize = uint8(textSizeValue - 1)
	winDecor.TitleSize = uint8(titleSizeValue - 1)

	uiWindow := cdashdisplay.UIWindow{
		Dims: cdashdisplay.UIDimensions{
			X0:     uint16(xValue),
			Y0:     uint16(yValue),
			Width:  uint16(widthValue),
			Height: uint16(heightValue),
		},
		Opts: cdashdisplay.UIWindowOpts{
			WinType:      cdashdisplay.WinTypeString, // NOTE: values aren't implemented yet
			ShowID:       showIDValue,
			PreviewValue: helper.B32(form.PreviewValue.GetText()),
		},
		Decor: winDecor,
		Title: helper.B32(form.Title.GetText()),
	}

	uiData := cdashdisplay.DesktopUIData{
		TelemetryField: telemField,
	}

	return &cdashdisplay.DesktopUIWindow{
		UIWindow: uiWindow,
		UIData:   uiData,
	}, nil
}

// createWindow is the function the callback for the "Create" button on the new window
// form
// - parses the contents of the new window form
// - sends that data to the correct device service
func (lc *LayoutController) createWindow() {
	window, err := lc.parseWindowFormData(*lc.LayoutToolView.LayoutActions.CreateWindowView.Form)
	if err != nil {
		lc.Messages <- "failed to parse form inputs: " + err.Error()
		return
	}

	lc.Messages <- fmt.Sprintf("pre  update w address: %p\n", window)
	window, err = lc.DevService.CreateWindow(window)
	if err != nil {
		lc.Messages <- "failed to create window\n"
		return
	}
	lc.Messages <- fmt.Sprintf("post update w address: %p\n", window)
	lc.Messages <- fmt.Sprintf("%+v\n", window)

	err = lc.updateFormView(window)
	if err != nil {
		// NOTE: if we fail to append the window to the views we must delete it
		// altogether given we won't be able to manipulate it any further
		lc.Messages <- "failed to append window to views " + err.Error()
	}
}

// updateFormView updates the new window form to be an existing window form and we change
// the buttons for that purpuse
func (lc *LayoutController) updateFormView(win *cdashdisplay.DesktopUIWindow) error {
	// OnSuccess we update our form to be an existing window form

	lc.Messages <- fmt.Sprintf("Setting up window:\n%+v\n", win)

	telemFieldID, ok := telemetry.GetFieldID(win.UIData.TelemetryField)
	if !ok {
		telemFieldID = 1000
	}

	lc.Messages <- fmt.Sprintf("field id for: %s is %d\n", win.UIData.TelemetryField, telemFieldID)

	err := lc.LayoutToolView.WindowCreatedSuccessfuly(win)
	if err != nil {
		return err
	}

	// Set the update window button behaviour
	formView := lc.LayoutToolView.FormQuickAccess[win.UIData.IDX]
	lc.Messages <- fmt.Sprintf("QA FORM: %+v\n", lc.LayoutToolView.FormQuickAccess)

	err = SetFormButtonCallback(formView.Form.Form, "Update", func() {
		lc.Messages <- "pressed update form button\n"
		window, err := lc.parseWindowFormData(*formView.Form)
		if err != nil {
			lc.Messages <- "failed to parse window form data " + err.Error() + "\n"
			return
		}
		lc.Messages <- fmt.Sprintf("window data in form: %v\n", window)

		window.UIData.IDX = formView.WinID

		lc.updateWindowAction(window)
	})
	if err != nil {
		lc.Messages <- "failed to set callback for update button\n"
		return err
	}

	formView.Form.Form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEscape:
			lc.App.SetFocus(lc.LayoutToolView.LayoutTree.Tree)
		}

		return ev
	})

	lc.App.SetFocus(formView.Form.Form)

	return nil
}

func (lc *LayoutController) newWindowAction() {
	if lc.LayoutToolView.LayoutActions.CreateWindowView != nil {
		// already exists, just change focus and leave
		lc.App.SetFocus(lc.LayoutToolView.LayoutActions.CreateWindowView.Form.Form)
		lc.Messages <- "already exists"
		return
	}

	// Get the new window form view
	lc.LayoutToolView.LayoutActions.CreateWindowView = views.NewCreateWindowFormView()
	newWindowForm := lc.LayoutToolView.LayoutActions.CreateWindowView

	// Set the event capture for this form
	newWindowForm.Form.Form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEscape:
			lc.App.SetFocus(lc.LayoutToolView.LayoutTree.Tree)
		}

		return ev
	})

	// Set the behaviour for when we press the create button
	err := SetFormButtonCallback(newWindowForm.Form.Form, "Create", func() {
		lc.createWindow()
	})
	if err != nil {
		lc.Messages <- "Failed to set the callback for the create window button"
	}

	// Set this on the action page and change to it
	lc.LayoutToolView.ShowCreateWindowForm()

	// Set its focus
	lc.App.SetFocus(newWindowForm.Form.Form)
}

func (lc *LayoutController) updateWindowAction(win *cdashdisplay.DesktopUIWindow) {
	err := lc.DevService.UpdateWindow(win)

	lc.Messages <- fmt.Sprintf("Window: %v\n", win)

	if err != nil {
		lc.Messages <- "failed to update window due to " + err.Error() + "\n"
		return
	}
}

func (lc *LayoutController) displayLoadedLayouts() {
	for _, w := range lc.DevService.CDash.State.Layout.Windows {
		err := lc.updateFormView(w)
		if err != nil {
			lc.Messages <- "failed to add window to list"
		}
	}
}

func (lc *LayoutController) getCurrentTreeNodeModel() (*tview.TreeNode, int16, error) {
	// Grab the form for the currently selected window
	node := lc.LayoutToolView.LayoutTree.Tree.GetCurrentNode()
	if node == nil {
		return nil, -1, fmt.Errorf("failed to grab current tree node")
	}

	// From this node get its references
	ref, ok := node.GetReference().(int16)
	if !ok {
		return nil, -1, fmt.Errorf("failed to get data on currently selected window")
	}

	return node, ref, nil
}

func (lc *LayoutController) loadLayout() {
	// We would get the layout path from somewhere but for nots its layout.yaml
	err := lc.DevService.LoadLayout("layout.yaml")
	if err != nil {
		lc.Messages <- "failed to load layout: " + err.Error()
		return
	}

	lc.displayLoadedLayouts()
}

func (lc *LayoutController) saveLayout() {
	err := lc.DevService.SaveLayout("layout.yaml")
	if err != nil {
		lc.Messages <- "failed to save layout: " + err.Error()
		return
	}
}

// deleteWindow will delete the currently highlighted window in the layoutTree
// - get the currently highlighted element
// - tell the LayoutToolView too delete it
func (lc *LayoutController) deleteWindow() {
	// Delete the window first
	// Get the selected node info
	curNode := lc.LayoutToolView.LayoutTree.Tree.GetCurrentNode()
	if curNode == nil {
		lc.Messages <- "couldn't get a hold of currently selected node\n"
		return
	}
	wID := curNode.GetReference().(int16)

	// Delete it
	err := lc.DevService.DeleteWindow(wID)
	if err != nil {
		lc.Messages <- "failed to delete window: " + err.Error() + "\n"
		return
	}

	// Reflect the deletion in the view
	lc.LayoutToolView.DeleteWindowByNode(curNode)
}

// moveWindow is a utility to interactively move the windows on the display
func (lc *LayoutController) moveWindow() {
	lc.Messages <- "Entering move mode (ESC to exit)\n"

	_, idx, err := lc.getCurrentTreeNodeModel()
	if err != nil {
		lc.Messages <- "failed to enter move mode: " + err.Error() + "\n"
		return
	}

	formView := lc.LayoutToolView.FormQuickAccess[idx].Form
	oldCapture := formView.Form.GetInputCapture()

	formView.Form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		lc.Messages <- "Captured event??\n"
		switch ev.Key() {
		case tcell.KeyESC:
			formView.Form.SetInputCapture(oldCapture)
			lc.App.SetFocus(lc.LayoutToolView.LayoutTree.Tree)
		}

		// Mode switching
		switch ev.Rune() {
		case 'r':
			lc.MoveToolState.Mode = resizeMode
			lc.Messages <- "Switching to resizeMode"
			return nil
		case 'm':
			lc.MoveToolState.Mode = moveMode
			lc.Messages <- "Switching to moveMode"
			return nil
		}

		// Delegate the current handler input
		handler := lc.MoveToolState.CurrentHandler(lc, idx)
		if handler != nil {
			return handler(ev)
		}

		return nil
	})

	lc.App.SetFocus(formView.Form)
}
