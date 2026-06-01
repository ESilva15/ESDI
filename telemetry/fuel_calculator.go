package telemetry

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
)

// NOTE: for managing fuel consumption and predictions we need to filter out
// abnormal laps like going into the pits and refueling and whatnot

// NOTE: create a log file specifically for this, I reckon it would be better

type fuelCalcState int

const (
	MaxLaps = 256
)

const (
	// leaves pits or whenever this starts
	OutlapState fuelCalcState = iota
	// we cross the line for the first time, we can count real time usage here
	// but don't have history yet
	FirstLapState
	// We already have history and can therefore do averages
	NormalState
)

type lapFuelData struct {
	Lap            int     // Number of the lap of this data
	StartFuelLevel float64 // Fuel level at the start/finish line
	FuelUsage      float64 // Fuel used during the lap
	Valid          bool    // If we should consider this data for calculations
}

func (lfd *lapFuelData) SetLap(lap int) {
	lfd.Lap = lap
}

func (lfd *lapFuelData) SetStartFuelLevel(fuelLevel float64) {
	lfd.StartFuelLevel = fuelLevel
}

func (lfd *lapFuelData) InvalidateLap() {
	lfd.Valid = false
}

func (lfd *lapFuelData) CalculateFuelUsage(fuelLevel float64) {
	lfd.FuelUsage = lfd.StartFuelLevel - fuelLevel
}

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

	Laps       [MaxLaps]lapFuelData
	CurrentLap int

	averageFuelUsage float64
	expectedLaps     float64
}

func NewFuelCalculator(logger *slog.Logger) *FuelCalculator {
	return &FuelCalculator{
		Logger:           logger,
		setup:            &sync.Once{},
		lapHistory:       make([]float64, 0, 5), // Pre-allocate space for 5 laps
		maxHistoryLaps:   3,                     // We want a 3-lap rolling average
		lastLapNumber:    0,
		state:            OutlapState,
		CurrentLap:       0,
		averageFuelUsage: 0.0,
	}
}

func (fc *FuelCalculator) initializeFCTelemetryFields(td *TelemetryData) {
	td.Values[FCLastLap].Type = DataTypeSTRING
	td.Values[FCLastLap].Str = "No Data"

	td.Values[FCCurrentLap].Type = DataTypeSTRING
	td.Values[FCCurrentLap].Str = "No Data"

	td.Values[FCExpectedLaps].Type = DataTypeSTRING
	td.Values[FCExpectedLaps].Str = "No Data"

	td.Values[FCAverage].Type = DataTypeSTRING
	td.Values[FCAverage].Str = "No Data"
}

func (fc *FuelCalculator) newLapData(lap int, fuelLevel float64) {
	fc.Laps[fc.CurrentLap].SetStartFuelLevel(fuelLevel)
	fc.Laps[fc.CurrentLap].SetLap(lap)
	fc.Laps[fc.CurrentLap].Valid = true
}

func (fc *FuelCalculator) MaxLapsReached(td *TelemetryData) {
	td.Values[FCLastLap].Str = "MAX!!!"
	td.Values[FCCurrentLap].Str = "MAX!!!"
}

func (fc *FuelCalculator) updateLapFuelUsage(td *TelemetryData, fuel float64) {
	fc.Logger.Debug(fmt.Sprintf("= %-2d =============================================", fc.CurrentLap))
	fc.Logger.Debug(fmt.Sprintf("-> Updating Lap Fuel Usage"))
	fc.Logger.Debug(fmt.Sprintf("   %+v", fc.Laps[fc.CurrentLap]))
	fc.Logger.Debug(fmt.Sprintf("   %f - %f = %f",
		fc.Laps[fc.CurrentLap].StartFuelLevel, fuel, fc.Laps[fc.CurrentLap].StartFuelLevel-fuel))

	fc.Laps[fc.CurrentLap].CalculateFuelUsage(fuel)
	td.Values[FCLastLap].Str = fmt.Sprintf("%.1f", fc.Laps[fc.CurrentLap].FuelUsage)
}

func (fc *FuelCalculator) updateAverageFuelUsage(td *TelemetryData) {
	// NOTE: this is a dumb approach on how to calculte fuel usage
	// sum := 0.0
	// lapCount := 0
	// for k := 0; k < fc.CurrentLap; k++ {
	// 	if !fc.Laps[k].Valid {
	// 		continue
	// 	}
	//
	// 	sum += fc.Laps[k].FuelUsage
	// 	lapCount += 1
	// }
	//
	// if lapCount == 0 {
	// 	return
	// }
	//
	// fc.averageFuelUsage = sum / float64(lapCount)
	fc.averageFuelUsage = fc.Laps[fc.CurrentLap].FuelUsage

	td.Values[FCAverage].Str = fmt.Sprintf("%.2f", fc.averageFuelUsage)
}

func (fc *FuelCalculator) updateExpectedLaps(td *TelemetryData, fuel float64) {
	fc.expectedLaps = fuel / fc.averageFuelUsage

	td.Values[FCExpectedLaps].Str = fmt.Sprintf("%.1f", fc.expectedLaps)
}

func (fc *FuelCalculator) startFinishLineCrossed(td *TelemetryData, lap int, fuel float64) {
	if fc.state == OutlapState {
		fc.state = FirstLapState
	}

	if fc.state == FirstLapState {
		fc.state = NormalState
	}

	// if fc.state == NormalState {
	// 	// In the normal state we have normal calculations to compute
	// 	fc.updateLapFuelUsage(td, fuel)
	// }

	if fc.CurrentLap == MaxLaps {
		fc.MaxLapsReached(td)
		return
	}

	// 1. Update last lap (we are still on the curret index) lap's fuel usage
	fc.updateLapFuelUsage(td, fuel)
	// 2. Update the average usage
	fc.updateAverageFuelUsage(td)
	// 3. Update the expected laps
	fc.updateExpectedLaps(td, fuel)

	// Go into the next lap
	fc.CurrentLap += 1
	fc.newLapData(lap, fuel)
}

func (fc *FuelCalculator) updateCurrentLapFuelUsage(td *TelemetryData, fuel float64) {
	// NOTE: find a more performant way of float64 -> string
	fuelDelta := fc.Laps[fc.CurrentLap].StartFuelLevel - fuel

	td.Values[FCCurrentLap].Str = fmt.Sprintf("%.1f", fuelDelta)
}

func (fc *FuelCalculator) updateRealTimeVariables(td *TelemetryData, lap int, fuel float64) {
	// if fc.state == OutlapState {
	// Nothing to calculate here currently
	// 	return
	// }

	// Real time updates
	fc.updateCurrentLapFuelUsage(td, fuel)
}

// NOTE: I need to add better logic for this. We need to detect pit stops
// instead of just checking if the fuel is greater than before
func (fc *FuelCalculator) refueled(fuel float64) bool {
	return fuel > fc.Laps[fc.CurrentLap].StartFuelLevel
}

func (fc *FuelCalculator) Process(td *TelemetryData) {
	// This setup function will set the default values for the fields we work
	// with on this fuel calculator
	fc.setup.Do(func() {
		fc.initializeFCTelemetryFields(td)
	})

	scannedLap := int(td.Values[LapNumber].Raw)
	// NOTE: this is stupid as, fix it. I mean don't store fuel level as a string
	scannedFuelLevel, err := strconv.ParseFloat(td.Values[FuelLevel].Str, 64)
	if err != nil {
		fc.Logger.Debug(fmt.Sprintf("error parsing fuel level: %+v", err))

		td.Values[FCLastLap].Str = "ERROR"
		td.Values[FCCurrentLap].Str = "ERROR"

		return
	}

	// Start the 0th lap
	if fc.CurrentLap == 0 && fc.Laps[0].StartFuelLevel == 0 {
		fc.newLapData(scannedLap, scannedFuelLevel)
	}

	if fc.refueled(scannedFuelLevel) {
		fc.Laps[fc.CurrentLap].StartFuelLevel = scannedFuelLevel
		fc.Laps[fc.CurrentLap].InvalidateLap()
	}

	if scannedLap > fc.CurrentLap {
		// Crossed the start finish line to the next lap
		fc.startFinishLineCrossed(td, scannedLap, scannedFuelLevel)
	}

	fc.updateRealTimeVariables(td, scannedLap, scannedFuelLevel)
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
