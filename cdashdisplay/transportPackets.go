package cdashdisplay

// In this file we will place all structs that are 1:1 representation of the
// types in the device transport layer ->

const (
	ShowIDFalse uint8 = 0
	ShowIDTrue  uint8 = 1
)

const (
	WinTypeString uint8 = 0
)

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

// NOTE: we can't use FString32 for this - too many bytes
// NOTE: use a bit flags for this options instead
type UIWindowOpts struct {
	ShowID       uint8     `yaml:"ShowID"`
	WinType      uint8     `yaml:"WinType"`
	PreviewValue FString32 `yaml:"PreviewValue"`
}

type UIWindow struct {
	Dims  UIDimensions  `yaml:"Dims"`
	Decor UIDecorations `yaml:"Decor"`
	Opts  UIWindowOpts  `yaml:"Opts"`
	Title FString32     `yaml:"Title"`
}

type DesktopUIWindow struct {
	UIWindow
	UIData DesktopUIData
}

type DesktopUIData struct {
	IDX            int16  `yaml:"WID"`
	TelemetryField string `yaml:"TelemetryField"`
}

type UIWindowUpdatePacket struct {
	WinID  int16
	Window UIWindow
}

// func getUIWindowDTO(w *DesktopWindowData) *UIWindow {
//
// }
