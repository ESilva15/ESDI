package controllers

import (
	"esdi/devices/cdashdisplay"
	helper "esdi/helpers"

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

func (lc *LayoutController) handleMovementCapture(idx int16,
	ev *tcell.EventKey,
) *tcell.EventKey {
	vec, ok := keyToVector(ev.Rune())
	if !ok {
		return nil
	}

	// Acquire the cdashdisplay
	displayIF, err := lc.DevService.GetDevice(cdashdisplay.NAME)
	if err != nil {
		lc.Messages <- "failed to get " + cdashdisplay.NAME
		return nil
	}
	display, ok := displayIF.(*cdashdisplay.CDashDisplay)
	if !ok {
		lc.Messages <- "failed to acquire " + cdashdisplay.NAME
		return nil
	}
	// ---

	err = display.MoveWindow(idx, &vec)
	if err != nil {
		lc.Messages <- "failed to move window: " + err.Error() + "\n"
		return nil
	}

	// Success - update the form
	window := display.State.Layout.Windows[idx]
	lc.LayoutToolView.UpdateFormView(idx, window)

	return nil
}

func (lc *LayoutController) handleResizeCapture(idx int16,
	ev *tcell.EventKey,
) *tcell.EventKey {
	vec, ok := keyToVector(ev.Rune())
	if !ok {
		return nil
	}

	// Acquire the cdashdisplay
	displayIF, err := lc.DevService.GetDevice(cdashdisplay.NAME)
	if err != nil {
		lc.Messages <- "failed to get " + cdashdisplay.NAME
		return nil
	}
	display, ok := displayIF.(*cdashdisplay.CDashDisplay)
	if !ok {
		lc.Messages <- "failed to acquire " + cdashdisplay.NAME
		return nil
	}
	// ---

	err = display.ResizeWindow(idx, &vec)
	if err != nil {
		lc.Messages <- "failed to resize window: " + err.Error() + "\n"
		return nil
	}

	// Success - update the form
	window := display.State.Layout.Windows[idx]
	lc.LayoutToolView.UpdateFormView(idx, window)

	return nil
}

func (s *windowManipState) CurrentHandler(lc *LayoutController, idx int16) modeHandler {
	switch s.Mode {
	case moveMode:
		return func(ev *tcell.EventKey) *tcell.EventKey {
			return lc.handleMovementCapture(idx, ev)
		}
	case resizeMode:
		return func(ev *tcell.EventKey) *tcell.EventKey {
			return lc.handleResizeCapture(idx, ev)
		}
	}

	return nil
}
