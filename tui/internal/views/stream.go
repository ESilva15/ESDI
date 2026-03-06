package views

import (
	telem "esdi/telemetry"
	"fmt"
	"strings"
	"time"

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

func (sv *StreamView) Update(data *telem.TelemetryData) {
	sv.TextView.SetText(stringify(data))
}

func stringify(data *telem.TelemetryData) string {
	var buffer strings.Builder

	delta := data.LastDataPoll.Sub(data.PenultimateDataPoll)
	buffer.WriteString(fmt.Sprintf("[%s]\n", time.Now().Format("2006/01/02 15:04:05.000")))
	buffer.WriteString(fmt.Sprintf("Delta: %d [%f]\n\n", delta.Milliseconds(), 1000.0/60.0))

	buffer.WriteString(fmt.Sprintf("Gear: %s, RPM: %s, Speed: %s\n",
		data.Values[telem.Gear].String(),
		data.Values[telem.RPM].String(),
		data.Values[telem.Speed].String(),
	))

	return buffer.String()
}
