package views

import (
	"fmt"
	"strings"
	"time"

	"esdi/providers"
	telem "esdi/telemetry"

	"github.com/rivo/tview"
)

// Car data lengths
const (
	SpeedLen     = 5
	GearLen      = 3
	RpmLen       = 6
	BrakeBiasLen = 6
)

// Form to select the game and whatnot ↓↓↓↓

// StreamOptionsView will be our view element to show the stream data
type StreamOptionsView struct {
	Form        *tview.Form
	SimDropdown *tview.DropDown
	UpdateBtn   *tview.Button
}

func NewStreamOptionsView(providerList []providers.Provider, defaultProvider string,
) *StreamOptionsView {
	sov := &StreamOptionsView{}

	sov.Form = tview.NewForm()
	sov.Form.SetTitle("Stream Options").SetBorder(true)

	defaultIdx := 0
	sov.SimDropdown = tview.NewDropDown().SetLabel("SIM").SetCurrentOption(0)
	for k, prov := range providerList {
		sov.SimDropdown.AddOption(prov.Name, func() {})

		if prov.Name == defaultProvider {
			defaultIdx = k
		}
	}
	sov.SimDropdown.SetCurrentOption(defaultIdx)
	sov.Form.AddFormItem(sov.SimDropdown)

	// Inject callback on the controller
	sov.Form.AddButton("Update", func() {})

	return sov
}

// Form to select the game and whatnot ↑↑↑↑

// Area to visualize what data is being passed to the game and whatnot ↓↓↓↓

type StreamVisualizerView struct {
	TextView *tview.TextView
}

func NewStreamVisualizerView() *StreamVisualizerView {
	tv := tview.NewTextView()
	tv.SetTitle("Streaming Visualizer").SetBorder(true)

	return &StreamVisualizerView{
		TextView: tv,
	}
}

func (sv *StreamVisualizerView) Update(data *telem.TelemetryData) {
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

// Area to visualize what data is being passed to the game and whatnot ↑↑↑↑

// Stream Tool ↓↓↓↓

type StreamToolView struct {
	Flex       *tview.Flex
	Options    *StreamOptionsView
	Visualizer *StreamVisualizerView
}

func NewStreamToolView(providerList []providers.Provider, defaultProvider string) *StreamToolView {
	optionsView := NewStreamOptionsView(providerList, defaultProvider)
	visualizerView := NewStreamVisualizerView()
	flex := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex.SetTitle("Streaming Tool")

	flex.
		AddItem(optionsView.Form, 0, 2, true).
		AddItem(visualizerView.TextView, 0, 5, true)

	return &StreamToolView{
		Flex:       flex,
		Options:    optionsView,
		Visualizer: visualizerView,
	}
}

// Stream Tool ↑↑↑↑
