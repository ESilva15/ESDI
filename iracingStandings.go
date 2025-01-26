package main

import (
	"sort"

	"github.com/ESilva15/goirsdk"
)

type StandingsLine struct {
	CarIdx     int
	LapPct     float32
	Lap        int32
	DriverName string
	EstTime    float32
	TimeBehind float32
}

type standingsFilter func([]StandingsLine, []float32, int)

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
	}
}

func relativeStandings(i *goirsdk.IBT, s []StandingsLine, id int) {
	// We sort the racists solely by their position on the track because all
	// we want to know is where people are in relation to us
	// But we should add a way to diferentatiate lap counts
	sort.Slice(s, func(i int, j int) bool {
		return s[i].LapPct > s[j].LapPct
	})

  estTime := i.Vars.Vars["CarIdxEstTime"].Value.([]float32)

	// Get the delta to a given carId
	for p := range s {
		curCarEstimate := estTime[s[p].CarIdx]

		var delta float32 = 0.0
		delta = abs(estTime[id] - curCarEstimate)

		s[p].TimeBehind = delta
	}
}

func bestLapTime(i *goirsdk.IBT, id int) float32 {
	best := i.Vars.Vars["LapBestLapTime"].Value.(float32)

	if best > 0 {
		return best
	}

	return float32(i.SessionInfo.DriverInfo.Drivers[id].CarClassEstLapTime)
}

// createStandingsTable will create a table with the standings data
// still working on this
// If the filter applied is generalStandings, the carId has to be 0
// We can do some dynamicProgramming on this thing I guess, I still haven't
// though about it much yet tbh
func createStandingsTable(i *goirsdk.IBT) []StandingsLine {
	driversLapDistPct := i.Vars.Vars["CarIdxLapDistPct"].Value.([]float32)
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
			CarIdx:     k,
			LapPct:     driversLapDistPct[k],
			DriverName: drivers[k].UserName,
			EstTime:    driversEstTime[k],
			Lap:        driversLap[k],
			TimeBehind: 0,
		}
	}

	return standings
}

func abs(v float32) float32 {
	if v < 0 {
		v = v * -1
	}
	return v
}



func getPlayerPosition(s []StandingsLine, p int) {
}
