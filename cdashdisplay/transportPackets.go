package cdashdisplay

// In this file we will place all structs that are 1:1 representation of the
// types in the device transport layer ->

type UIDimensions struct {
	X0     uint16 `yaml:"X0"`
	Y0     uint16 `yaml:"Y0"`
	Width  uint16 `yaml:"Width"`
	Height uint16 `yaml:"Height"`
}

type UIDecorations struct {
	BGColour     uint16 `yaml:"BGColour"`
	FGColour     uint16 `yaml:"FGColour"`
	TitleColour  uint16 `yaml:"TitleColour"`
	BorderColour uint16 `yaml:"BorderColour"`
	TitleSize    uint8  `yaml:"TitleSize"`
	TextSize     uint8  `yaml:"TextSize"`
	HasBorder    uint8  `yaml:"HasBorder"`
	Padding      uint8  `yaml:"Padding"`
}

type UIWindow struct {
	Dims  UIDimensions  `yaml:"Dims"`
	Decor UIDecorations `yaml:"Decor"`
	Title FString32     `yaml:"Title"`
}

type UIWindowUpdatePacket struct {
	WinID  int16
	Window UIWindow
}
