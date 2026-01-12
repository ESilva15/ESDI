package events

import (
	"esdi/tui/internal/models"
)

type Error struct {
	Error error
}

type WindowCreated struct {
	Window models.Window
}
