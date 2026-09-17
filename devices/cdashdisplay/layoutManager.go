package cdashdisplay

import (
	"fmt"
	"log/slog"
)

type LayoutTree struct {
	Windows map[int16]*DesktopUIWindow `yaml:"Windows"`
}

func NewLayoutTree() *LayoutTree {
	return &LayoutTree{
		Windows: make(map[int16]*DesktopUIWindow),
	}
}

func (l *LayoutTree) AddWindow(w *DesktopUIWindow) {
	slog.Debug(fmt.Sprintf("adding window '%d' - %v", w.UIData.IDX))
	l.Windows[w.UIData.IDX] = w
	slog.Debug(fmt.Sprintf("new map - %v", l.Windows))
}

func (l *LayoutTree) RemoveWindow(idx int16) {
	slog.Debug(fmt.Sprintf("removing window '%d'", idx))
	delete(l.Windows, idx)
	slog.Debug(fmt.Sprintf("new map - %v", l.Windows))
}
