package tviewhelpers

import "github.com/rivo/tview"

func CreateModal(
	background tview.Primitive,
	popup tview.Primitive,
	width int,
	height int,
) *tview.Pages {
	modal := func(p tview.Primitive, mWidth, mHeight int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(
				tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(nil, 0, 1, false).
					AddItem(p, mHeight, 1, true).
					AddItem(nil, 0, 1, false), mWidth, 1, true,
			).
			AddItem(nil, 0, 1, false)
	}

	pages := tview.NewPages().
		AddPage("background", background, true, true).
		AddPage("modal", modal(popup, width, height), true, true)

	return pages
}
