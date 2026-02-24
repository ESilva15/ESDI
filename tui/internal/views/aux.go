package views

import (
	"esdi/tui/internal/dom"
	"slices"

	"github.com/rivo/tview"
)

func AddAndShowPage(
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

	return nil
}

// func PageExists(pages *tview.Pages, needle string) bool {
// 	return slices.Contains(pages.GetPageNames(false), needle)
// }
