// Package t
package t

import "esdi/t/internal/tui"

func Run() error {
	app := tui.NewTUI()
	return app.Start()
}
