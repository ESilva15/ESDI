package views

import (
	"esdi/telemetry"
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// Car data lengths
const (
	SpeedLen     = 5
	GearLen      = 3
	RpmLen       = 6
	BrakeBiasLen = 6
)

type StreamView struct {
	TextView *tview.TextView
}

func NewStreamView() *StreamView {
	tv := tview.NewTextView()
	tv.SetTitle("Streaming Tool").SetBorder(true)

	return &StreamView{
		TextView: tv,
	}
}

func (sv *StreamView) Update(data telemetry.TelemetryData) {
	sv.TextView.SetText(stringify(data))
}

func stringify(data telemetry.TelemetryData) string {
	var buffer strings.Builder

	buffer.WriteString(fmt.Sprintf("Gear: %d, RPM: %d, Speed: %d\n\n",
		data.Values["Gear"], data.Values["RPM"], data.Values["Speed"]))

	return buffer.String()
}
