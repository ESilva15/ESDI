// Package ui
package ui

import "github.com/rivo/tview"

type UIContext struct {
	Redraw      func()
	Log         func(formatter string, args ...any)
	ChangeFocus func(p tview.Primitive)
}
