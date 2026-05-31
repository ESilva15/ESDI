package telemetry

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
)

// NOTE: for managing fuel consumption and predictions we need to filter out
// abnormal laps like going into the pits and refueling and whatnot

type fuelCalcState int

const (
	// leaves pits or whenever this starts
	outlap fuelCalcState = iota
	// we cross the line for the first time, we can count real time usage here
	// but don't have history yet
	firstLap
	// We already have history and can therefore do averages
	normal
)

// FuelCalculator acts as a Virtual Field middleware
type FuelCalculator struct {
	Logger *slog.Logger

	// --- Internal State (The Memory) ---
	lastLapNumber  int     // Tracks when we cross the start/finish line
	fuelAtLapStart float64 // Snapshot of fuel when the lap began

	setup *sync.Once

	lapHistory     []float64 // Buffer holding the last N valid laps
	maxHistoryLaps int       // How many laps to keep in the buffer (e.g., 3)

	// --- Validation Flags ---

	isOutLap       bool // True if we just left the pits
	isUnderCaution bool // True if safety car is out (optional, based on flags)

	// --- The Outputs (What gets pushed back to the TUI) ---

	lastLapUsage float64
	avgUsage     float64
	maxUsage     float64

	state fuelCalcState
}

func NewFuelCalculator(logger *slog.Logger) *FuelCalculator {
	return &FuelCalculator{
		Logger:         logger,
		setup:          &sync.Once{},
		lapHistory:     make([]float64, 0, 5), // Pre-allocate space for 5 laps
		maxHistoryLaps: 3,                     // We want a 3-lap rolling average
		lastLapNumber:  0,
		state:          outlap,
	}
}

func (fc *FuelCalculator) transitionState(state fuelCalcState) {
	fc.state = state
}

func (fc *FuelCalculator) crossedStartFinishLine(lap int) bool {
	return lap > fc.lastLapNumber
}

func (fc *FuelCalculator) refueled(currentFuelLevel float64) bool {
	return currentFuelLevel > fc.fuelAtLapStart
}

func (fc *FuelCalculator) storeStartFinishLineFuelLevels(curFuel float64, lap int) {
	fc.fuelAtLapStart = curFuel
	fc.lastLapNumber = lap
}

func (fc *FuelCalculator) calculateCurLapFuelUsage(td *TelemetryData, fuel float64) {
	td.Values[FuelCurrentLap].Type = DataTypeSTRING
	td.Values[FuelCurrentLap].Str = fmt.Sprintf("%.2f", fc.fuelAtLapStart-fuel)
}

func (fc *FuelCalculator) processOutlap(td *TelemetryData, lap int) {
	currentFuelLevel, err := strconv.ParseFloat(td.Values[FuelLevel].Str, 64)
	if err != nil {
		fc.Logger.Debug(fmt.Sprintf("error parsing fuel level: %+v", err))
		return
	}

	if fc.crossedStartFinishLine(lap) {
		fc.storeStartFinishLineFuelLevels(currentFuelLevel, lap)
		fc.transitionState(firstLap)
	}

	if fc.refueled(currentFuelLevel) {
		fc.transitionState(outlap)
	}
}

func (fc *FuelCalculator) processFirstlap(td *TelemetryData, lap int) {
	currentFuelLevel, err := strconv.ParseFloat(td.Values[FuelLevel].Str, 64)
	if err != nil {
		fc.Logger.Debug(fmt.Sprintf("error parsing fuel level: %+v", err))
		return
	}

	if fc.crossedStartFinishLine(lap) {
		fc.storeStartFinishLineFuelLevels(currentFuelLevel, lap)
		fc.transitionState(normal)
	}

	if fc.refueled(currentFuelLevel) {
		fc.transitionState(outlap)
	}

	fc.calculateCurLapFuelUsage(td, currentFuelLevel)
}

func (fc *FuelCalculator) processNormal(td *TelemetryData, lap int) {
	currentFuelLevel, err := strconv.ParseFloat(td.Values[FuelLevel].Str, 64)
	if err != nil {
		fc.Logger.Debug(fmt.Sprintf("error parsing fuel level: %+v", err))
		return
	}

	if fc.crossedStartFinishLine(lap) {
		fc.storeStartFinishLineFuelLevels(currentFuelLevel, lap)
		fc.transitionState(normal)
	}

	if fc.refueled(currentFuelLevel) {
		fc.transitionState(outlap)
	}

	fc.calculateCurLapFuelUsage(td, currentFuelLevel)

	// And here we add how to calculate last lap usage or whatever yo
}

func (fc *FuelCalculator) Process(td *TelemetryData) {
	// This setup function will set the default values for the fields we work
	// with on this fuel calculator
	fc.setup.Do(func() {
		td.Values[FuelLastLap].Type = DataTypeSTRING
		td.Values[FuelLastLap].Str = "No Data"

		td.Values[FuelCurrentLap].Type = DataTypeSTRING
		td.Values[FuelCurrentLap].Str = "No Data"
	})

	currentLap := int(td.Values[LapNumber].Raw)
	// NOTE: this is stupid as, fix it. I mean don't store fuel level as a string
	// currentFuelLevel, err := strconv.ParseFloat(td.Values[FuelLevel].Str, 64)

	switch fc.state {
	case outlap:
		fc.processOutlap(td, currentLap)
	case firstLap:
		fc.processFirstlap(td, currentLap)
	case normal:
		fc.processNormal(td, currentLap)
	}
}

// Simple helper to wipe state on session changes
func (fc *FuelCalculator) resetHistory() {
	fc.lapHistory = make([]float64, 0, fc.maxHistoryLaps)
	fc.lastLapNumber = 0
	fc.avgUsage = 0
	fc.maxUsage = 0
}

func (fc *FuelCalculator) EnsureSubscribed() []FieldID {
	return []FieldID{FuelLevel, LapNumber}
}
