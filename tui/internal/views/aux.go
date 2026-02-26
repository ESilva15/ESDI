package views

import (
	"slices"

	"github.com/rivo/tview"
)

func AddAndShowPage(pages *tview.Pages, pageID string, page tview.Primitive) {
	if !slices.Contains(pages.GetPageNames(false), pageID) {
		pages.AddAndSwitchToPage(pageID, page, true)
	} else {
		// Only change to page
		pages.SwitchToPage(pageID)
	}
}
