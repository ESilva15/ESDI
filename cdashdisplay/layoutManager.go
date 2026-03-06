package cdashdisplay

import "fmt"

type LayoutTree struct {
	Windows map[int16]*UIWindow `yaml:"Windows"`
}

func NewLayoutTree() *LayoutTree {
	return &LayoutTree{
		Windows: make(map[int16]*UIWindow),
	}
}

func (l *LayoutTree) AddWindow(idx int16, w UIWindow) {
	pLogger.Debug(fmt.Sprintf("adding window '%d' - %v", idx, w))
	l.Windows[idx] = &w
	pLogger.Debug(fmt.Sprintf("new map - %v", l.Windows))
}

func (l *LayoutTree) RemoveWindow(idx int16) {
	pLogger.Debug(fmt.Sprintf("removing window '%d'", idx))
	delete(l.Windows, idx)
	pLogger.Debug(fmt.Sprintf("new map - %v", l.Windows))
}
