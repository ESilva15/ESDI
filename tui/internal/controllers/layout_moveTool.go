package controllers

import (
	helper "esdi/helpers"
	"esdi/tui/internal/models"

	"github.com/gdamore/tcell/v2"
)

type ManipMode int

const (
	moveMode ManipMode = iota
	resizeMode
)

type modeHandler func(*tcell.EventKey) *tcell.EventKey

type windowManipState struct {
	Mode ManipMode
}

func keyToVector(r rune) (helper.Vector, bool) {
	var mult uint16 = 1

	// Is uppercase? For fast movement
	if r >= 'A' && r <= 'Z' {
		mult = 10
		r += 'a' - 'A'
	}

	switch r {
	case 'h':
		return helper.Vector{DX: -mult, DY: 0}, true
	case 'j':
		return helper.Vector{DX: 0, DY: mult}, true
	case 'k':
		return helper.Vector{DX: 0, DY: -mult}, true
	case 'l':
		return helper.Vector{DX: mult, DY: 0}, true
	}

	return helper.Vector{}, false
}

func (lc *LayoutController) handleMovementCapture(win *models.UIWindow,
	ev *tcell.EventKey) *tcell.EventKey {
	vec, ok := keyToVector(ev.Rune())
	if !ok {
		return nil
	}

	err := lc.DevService.MoveWindow(win, &vec)
	if err != nil {
		lc.Messages <- "failed to move window: " + err.Error() + "\n"
		return nil
	}

	// Success - update the form
	lc.LayoutToolView.UpdateFormView(win)

	return nil
}

func (lc *LayoutController) handleResizeCapture(win *models.UIWindow,
	ev *tcell.EventKey) *tcell.EventKey {
	vec, ok := keyToVector(ev.Rune())
	if !ok {
		return nil
	}

	err := lc.DevService.ResizeWindow(win, &vec)
	if err != nil {
		lc.Messages <- "failed to resize window: " + err.Error() + "\n"
		return nil
	}

	// Success - update the form
	lc.LayoutToolView.UpdateFormView(win)

	return nil
}

func (s *windowManipState) CurrentHandler(lc *LayoutController, w *models.UIWindow) modeHandler {
	switch s.Mode {
	case moveMode:
		return func(ev *tcell.EventKey) *tcell.EventKey {
			return lc.handleMovementCapture(w, ev)
		}
	case resizeMode:
		return func(ev *tcell.EventKey) *tcell.EventKey {
			return lc.handleResizeCapture(w, ev)
		}
	}

	return nil
}
