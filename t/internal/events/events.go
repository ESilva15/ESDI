package events

import (
	"esdi/t/internal/models"
)

type Error struct {
	Error error
}

type WindowCreated struct {
	Window models.Window
}
