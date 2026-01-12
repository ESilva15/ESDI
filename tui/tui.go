// Package tui
package tui

import "esdi/tui/internal/tui"

func Run() error {
	app := tui.NewTUI()
	return app.Start()
}
