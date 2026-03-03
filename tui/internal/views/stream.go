package views

import "github.com/rivo/tview"

// Car data lengths
const (
	SpeedLen     = 5
	GearLen      = 3
	RpmLen       = 6
	BrakeBiasLen = 6
)

const (
	streamingBoxID = "streaming-box"
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

func (sv *StreamView) Update(str string) {
	sv.TextView.SetText(str)
}
