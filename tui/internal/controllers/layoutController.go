package controllers

import (
	"esdi/cdashdisplay"
	helper "esdi/helpers"
	"esdi/tui/internal/models"
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
}

func NewLayoutController(base *Controller, service *services.CDashService) *LayoutController {
	lc := &LayoutController{
		Controller:     base,
		LayoutToolView: views.NewLayoutToolView(),
		Messages:       make(chan string, 10),
		DevService:     service,
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

		nodeWinRef := nodeRef.(*models.UIWindow)
		lc.Messages <- fmt.Sprintf("changing to -> %d\n", nodeWinRef.WID)

		lc.LayoutToolView.ShowWindowFormByID(nodeWinRef.WID)
	})

	lc.LayoutToolView.LayoutTree.Tree.SetSelectedFunc(func(node *tview.TreeNode) {
		// Change focus to the selected window
		nodeRef := node.GetReference()
		if nodeRef == nil {
			// its probably root I guess, thats what I'll believe
			return
		}

		nodeWinRef := nodeRef.(*models.UIWindow)

		lc.App.SetFocus(lc.LayoutToolView.FormQuickAccess[nodeWinRef.WID].Form.Form)
	})
}

func (lc *LayoutController) parseWindowFormData(
	form views.CDashDisplayWindowFormView,
) (*cdashdisplay.UIWindow, error) {

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

	showIDValue := cdashdisplay.ShowIDFalse
	if form.ShowID.IsChecked() {
		showIDValue = cdashdisplay.ShowIDTrue
	}

	winDecor := cdashdisplay.DefaultDecorations
	winDecor.TextSize = uint8(textSizeValue - 1)
	winDecor.TitleSize = uint8(titleSizeValue - 1)

	uiWindow := &cdashdisplay.UIWindow{
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

	return uiWindow, nil
}

func (lc *LayoutController) createWindow() {
	window, err := lc.parseWindowFormData(*lc.LayoutToolView.LayoutActions.CreateWindowView.Form)
	if err != nil {
		lc.Messages <- "failed to parse form inputs: " + err.Error()
		return
	}

	wID, err := lc.DevService.CreateWindow(window)
	if err != nil {
		lc.Messages <- "failed to create window\n"
		return
	}

	err = lc.updateFormView(
		&models.UIWindow{
			WID:  wID,
			Data: *window,
		},
	)
	if err != nil {

		// NOTE: if we fail to append the window to the views we must delete it
		// altogether given we won't be able to manipulate it any further
		lc.Messages <- "failed to append window to views " + err.Error()
	}
}

func (lc *LayoutController) updateFormView(win *models.UIWindow) error {
	// OnSuccess we update our form to be an existing window form
	err := lc.LayoutToolView.WindowCreatedSuccessfuly(win)
	if err != nil {
		return err
	}

	// Set the update window button behaviour
	formView := lc.LayoutToolView.FormQuickAccess[win.WID]
	err = SetFormButtonCallback(formView.Form.Form, "Update", func() {
		lc.Messages <- "pressed update form button\n"
		window, err := lc.parseWindowFormData(*formView.Form)
		if err != nil {
			lc.Messages <- "failed to parse window form data " + err.Error() + "\n"
			return
		}
		lc.Messages <- fmt.Sprintf("window data in form: %v\n", window)

		if err != nil {
			lc.Messages <- "failed to parse form: " + err.Error() + "\n"
			return
		}

		lc.updateWindowAction(&models.UIWindow{
			WID:  win.WID,
			Data: *window,
		})
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

func (lc *LayoutController) updateWindowAction(win *models.UIWindow) {
	err := lc.DevService.UpdateWindow(win.WID, &win.Data)

	lc.Messages <- fmt.Sprintf("Window: %v\n", win)

	if err != nil {
		lc.Messages <- "failed to update window due to " + err.Error() + "\n"
		return
	}
}

func (lc *LayoutController) displayLoadedLayouts() {
	for idx, w := range lc.DevService.CDash.State.Layout.Windows {

		err := lc.updateFormView(&models.UIWindow{WID: idx, Data: *w})
		if err != nil {
			lc.Messages <- "failed to add window to list"
		}
	}
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
	win := curNode.GetReference().(*models.UIWindow)

	// Delete it
	err := lc.DevService.DeleteWindow(win.WID)
	if err != nil {
		lc.Messages <- "failed to delete window: " + err.Error() + "\n"
		return
	}

	// Reflect the deletion in the view
	lc.LayoutToolView.DeleteWindowByNode(curNode)
}
