// Package models
package models

type Window struct {
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
