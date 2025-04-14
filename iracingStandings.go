package main

import (
	"sort"

	"github.com/ESilva15/goirsdk"
)

func createDriverName(s string) [DriverNameLen]byte {
	var arr [DriverNameLen]byte

	if len(s) > 32 {
		s = s[:32]
	}

	copy(arr[:], s)
	return arr
}

// type standingsFilter func([]StandingsLine, []float32, int)

func generalStandings(i *goirsdk.IBT, s []StandingsLine, id int) {
	// We have to sort the racists by their lap count and position on the track
	// on the current lap
	sort.Slice(s, func(i int, j int) bool {
		if s[i].Lap == int32(s[j].Lap) {
			return s[i].LapPct > s[j].LapPct
		}
		return s[i].Lap > s[j].Lap
	})

	estTime := i.Vars.Vars["CarIdxEstTime"].Value.([]float32)

	// This will give the delta to the guy ahead
	// P1 0:00:000
	// P2 0:01:000 -> 1s from P1
	// P3 0:01:000 -> 1s from P2
	for p := range s {
		theirEstimate := estTime[s[p].CarIdx]
		var theThing float32 = 0.0
		if p != id {
			theThing = estTime[s[p-1].CarIdx] - theirEstimate
		}

		s[p].TimeBehind = theThing
		copy(s[p].TimeBehindString[:], string(lapTimeRepresentation(theThing,
			RelativeDeltaFormatStr)))
	}
}

func relativeStandings(i *goirsdk.IBT, s []StandingsLine, id int) {
	sort.Slice(s, func(i int, j int) bool {
		if s[i].Lap == int32(s[j].Lap) {
			return s[i].LapPct > s[j].LapPct
		}
		return s[i].Lap > s[j].Lap
	})

	estTime := i.Vars.Vars["CarIdxEstTime"].Value.([]float32)

	// Get the delta to a given carId
	// TODO: fix the delta when the car behind is still in the previous lap.
	for p := range s {
		curCarEstimate := estTime[s[p].CarIdx]

		var delta float32 = 0.0
		delta = abs(estTime[id] - curCarEstimate)

		s[p].TimeBehind = delta
		copy(s[p].TimeBehindString[:], string(lapTimeRepresentation(delta,
			RelativeDeltaFormatStr)))
	}
}

// createStandingsTable will create a table with the standings data
// still working on this
// If the filter applied is generalStandings, the carId has to be 0
// We can do some dynamicProgramming on this thing I guess, I still haven't
// though about it much yet tbh
func createStandingsTable(i *goirsdk.IBT) []StandingsLine {
	driversLapDistPctRaw := i.Vars.Vars["CarIdxLapDistPct"].Value
  if driversLapDistPctRaw == nil {
    return []StandingsLine{}
  }

	driversLapDistPct := driversLapDistPctRaw.([]float32)
	driversEstTime := i.Vars.Vars["CarIdxEstTime"].Value.([]float32)
	driversLap := i.Vars.Vars["CarIdxLap"].Value.([]int32)
	drivers := i.SessionInfo.DriverInfo.Drivers

	standings := make([]StandingsLine, len(drivers))

	for k := range len(drivers) {
		if drivers[k].CarIsPaceCar == 1 || drivers[k].IsSpectator == 1 || drivers[k].UserName == "" {
			continue
		}

		if driversLap[k] == -1 || drivers[k].UserName == "" {
			continue
		}

		standings[k] = StandingsLine{
			CarIdx:     int32(k),
			LapPct:     driversLapDistPct[k],
			DriverName: createDriverName(drivers[k].UserName),
			EstTime:    driversEstTime[k],
			Lap:        driversLap[k],
			TimeBehind: 0,
		}
	}

	return standings
}

func abs[V int32 | float32 | int](value V) V {
	if value < 0 {
		value = value * -1
	}

	return value
}
