package cdashdisplay

import "fmt"

type LayoutTree struct {
	Windows map[int16]*DesktopUIWindow `yaml:"Windows"`
}

func NewLayoutTree() *LayoutTree {
	return &LayoutTree{
		Windows: make(map[int16]*DesktopUIWindow),
	}
}

func (l *LayoutTree) AddWindow(w *DesktopUIWindow) {
	pLogger.Debug(fmt.Sprintf("adding window '%d' - %v", w.UIData.IDX))
	l.Windows[w.UIData.IDX] = w
	pLogger.Debug(fmt.Sprintf("new map - %v", l.Windows))
}

func (l *LayoutTree) RemoveWindow(idx int16) {
	pLogger.Debug(fmt.Sprintf("removing window '%d'", idx))
	delete(l.Windows, idx)
	pLogger.Debug(fmt.Sprintf("new map - %v", l.Windows))
}
