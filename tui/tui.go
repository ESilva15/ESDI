// Package tui
package tui

import (
	t "esdi/tui/internal/tui"
	"log/slog"
)

func Run(logger *slog.Logger) error {
	return t.Start(logger)
}
