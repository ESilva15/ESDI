// Package models
package models

import "esdi/cdashdisplay"

type WindowForm struct {
	X            uint16
	Y            uint16
	Width        uint16
	Height       uint16
	TitleSize    uint8
	TextSize     uint8
	Title        string
	Type         uint8
	ShowID       uint8
	PreviewValue string
}

type UIWindow struct {
	WID  int16
	Data cdashdisplay.UIWindow
}
