package views

import (
	"esdi/tui/internal/dom"
	"esdi/tui/internal/events"
	"esdi/tui/internal/ui"
	"slices"

	"github.com/rivo/tview"
)

func AddAndShowPage(
	bus *events.Bus,
	doc *dom.DOM,
	pages *tview.Pages,
	page *dom.UINode,
	changeFocus bool,
) error {
	if !slices.Contains(pages.GetPageNames(false), page.ID) {
		pages.AddAndSwitchToPage(page.ID, page.Self, true)
	} else {
		// Only change to page
		pages.SwitchToPage(page.ID)
	}

	// Set focus to our new tool
	if changeFocus {
		bus.Emit(ui.ChangeFocusEv{Target: page.Self})
	}

	return nil
}

// func PageExists(pages *tview.Pages, needle string) bool {
// 	return slices.Contains(pages.GetPageNames(false), needle)
// }
