package controllers

import "esdi/tui/internal/events"

type LayoutController struct {
	EvBus          *events.Bus
	LayoutToolView any
}

func (lc *LayoutController) registerEvents() {
}

func NewLayoutController(bus *events.Bus) *LayoutController {
	return &LayoutController{
		EvBus: bus,
	}
}
