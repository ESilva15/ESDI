package telemetry

import (
	"fmt"
	"log/slog"
	"strconv"
)

type fuelCalcState int

const (
	outlap fuelCalcState = iota
	firstLap
	normal
)

// FuelCalculator acts as a Virtual Field middleware
type FuelCalculator struct {
	Logger *slog.Logger

	// --- Internal State (The Memory) ---
	lastLapNumber  int     // Tracks when we cross the start/finish line
	fuelAtLapStart float64 // Snapshot of fuel when the lap began

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
		lapHistory:     make([]float64, 0, 5), // Pre-allocate space for 5 laps
		maxHistoryLaps: 3,                     // We want a 3-lap rolling average
		lastLapNumber:  0,
		state:          outlap,
	}
}

// func (fc *FuelCalculator) isOutlap(lap int) {
// 	if fc.isOutLap
// }

func (fc *FuelCalculator) Process(td *TelemetryData) {
	currentLap := int(td.Values[LapNumber].Raw)
	// NOTE: this is stupid as, fix it
	currentFuelLevel, err := strconv.ParseFloat(td.Values[FuelLevel].Str, 64)
	if err != nil {
		fc.Logger.Debug(fmt.Sprintf("error parsing fuel level: %+v", err))
		return
	}

	if currentLap == 0 {
		td.Values[FuelLastLap].Type = DataTypeSTRING
		td.Values[FuelLastLap].Str = "No Data"

		td.Values[FuelCurrentLap].Type = DataTypeSTRING
		td.Values[FuelCurrentLap].Str = "No Data"

		return
	}

	if currentLap > fc.lastLapNumber {
		fc.Logger.Debug(fmt.Sprintf("crossed the line: %+v -> %+v", currentLap, fc.lastLapNumber))
		fc.state = firstLap
		fc.lastLapNumber = currentLap
		fc.fuelAtLapStart = currentFuelLevel
	}

	td.Values[FuelLastLap].Type = DataTypeSTRING
	td.Values[FuelLastLap].Str = "No Data"

	td.Values[FuelCurrentLap].Type = DataTypeSTRING
	fc.Logger.Debug(fmt.Sprintf("calculate: %+v - %+v = %+v", fc.fuelAtLapStart,
		currentFuelLevel, fc.fuelAtLapStart-currentFuelLevel))
	td.Values[FuelCurrentLap].Str = fmt.Sprintf("%.2f", fc.fuelAtLapStart-currentFuelLevel)
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
