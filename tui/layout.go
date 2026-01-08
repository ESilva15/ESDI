package tui

import "github.com/rivo/tview"

type ui struct {
	Elements map[string]*view
}

func NewUI() *ui {
	return &ui{
		Elements: make(map[string]*view),
	}
}

type view struct {
	ID       string
	Title    string
	Root     tview.Primitive
	Children []*view
}
