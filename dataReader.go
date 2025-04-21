package main

import (
	"fmt"
	"time"
)

func (e *ESDI) getVehicleData() {
	mu.Lock()
	curGear := e.irsdk.Vars.Vars["Gear"].Value
	curRPM := e.irsdk.Vars.Vars["RPM"].Value
	curSpeed := e.irsdk.Vars.Vars["Speed"].Value
	curBrakeBias, ok := e.irsdk.Vars.Vars["dcBrakeBias"]
	if ok {
		copyBytes(e.dataPacket.BrakeBias[:], BrakeBiasLen,
			fmt.Sprintf("%.1f", curBrakeBias.Value.(float32)))
	}

	e.data.Gear = int32(curGear.(int))
	e.data.RPM = int32(curRPM.(float32))
	e.data.Speed = int32(msToKph(curSpeed.(float32)))
	mu.Unlock()
}

func (e *ESDI) fuelData() {
	mu.Lock()
	fLiters := e.irsdk.Vars.Vars["FuelLevel"].Value
	fPct := e.irsdk.Vars.Vars["FuelLevelPct"].Value

	curLap := e.irsdk.Vars.Vars["Lap"].Value.(int)
	fuelLevels[curLap] = fLiters.(float32)

	if curLap-2 < 0 {
		e.data.FuelPerLap = 0.0
	} else {
		e.data.FuelPerLap = fuelLevels[curLap-2] - fuelLevels[curLap-1]
	}

	e.data.FuelPct = float32(fPct.(float32)) * 100
	e.data.FuelLiters = float32(fLiters.(float32))
	e.data.FuelTotal = (100 * e.data.FuelLiters) / e.data.FuelPct
	mu.Unlock()
}

func (e *ESDI) lapData() {
	mu.Lock()
	currentLap := e.irsdk.Vars.Vars["Lap"].Value
	lapDistPct := e.irsdk.Vars.Vars["LapDistPct"].Value
	currentLapTime := e.irsdk.Vars.Vars["LapCurrentLapTime"].Value
	lapBestLapTime := e.irsdk.Vars.Vars["LapBestLapTime"].Value
	lapLastLapTime := e.irsdk.Vars.Vars["LapLastLapTime"].Value
	lapDeltaToBestLap := e.irsdk.Vars.Vars["LapDeltaToBestLap"].Value

	e.data.LapDeltaFloat = lapDeltaToBestLap.(float32)
	e.data.LapCount = int32(currentLap.(int))
	e.data.LapDistPct = float32(lapDistPct.(float32)) * 100

	// TODO
	// Don't create the strings here, should be creating them later one only
	// Get the best lap data from the session info - I guess
	copy(e.data.CurrLapTime[:], string(lapTimeRepresentation(currentLapTime.(float32),
		LapTimeFormatStr)))
	copy(e.data.LastLapTime[:], string(lapTimeRepresentation(lapLastLapTime.(float32),
		LapTimeFormatStr)))
	copy(e.data.BestLapTime[:], string(lapTimeRepresentation(lapBestLapTime.(float32),
		LapTimeFormatStr)))
	mu.Unlock()
}

// This function doesnt feel very pretty - NEED TO REFACTOR IT
func (e *ESDI) positionData() {
	// Only create a table with more people if we have more players
	mu.Lock()
	standings := createStandingsTable(e.irsdk)
	if len(standings) <= 0 {
		for range 5 {
			standings = append(standings, paddingStandingsLine)
		}
	} else {
		relativeStandings(e.irsdk, standings, e.irsdk.SessionInfo.DriverInfo.DriverCarIdx)
		p := findEntry(standings, func(l StandingsLine) bool {
			return l.CarIdx == int32(e.irsdk.SessionInfo.DriverInfo.DriverCarIdx)
		})

		lowerLim := p - 2
		upperLim := p + 3

		var lowerPadding []StandingsLine
		var upperPadding []StandingsLine

		if lowerLim < 0 {
			lowerPadding = make([]StandingsLine, abs(lowerLim))
			for k := range abs(lowerLim) {
				lowerPadding[k] = paddingStandingsLine
			}
			lowerLim = 0
		}
		if upperLim >= len(standings) {
			upperPadding = make([]StandingsLine, upperLim-len(standings))
			for k := range upperLim - len(standings) {
				upperPadding[k] = paddingStandingsLine
			}
			upperLim = len(standings)
		}

		standings = append(lowerPadding, standings[lowerLim:upperLim]...)
		standings = append(standings, upperPadding...)

		e.data.Position = int32(p)
	}

	copy(e.data.Standings[:], standings[0:5])
	mu.Unlock()
}

func packageStandingsLineDataPacket(data [5]StandingsLine) [5]StandingsLineDataPacket {
	var sl [5]StandingsLineDataPacket

	for k := range data {
		copyBytes(sl[k].Lap[:], LapStringLen, fmt.Sprintf("%-2d", data[k].Lap))
		copyBytes(sl[k].DriverName[:], DriverNameLen,
			fmt.Sprintf("%-16s", data[k].DriverName[0:DriverNameLen]))
		copyBytes(sl[k].TimeBehindString[:], TimeBehindStringLen,
			fmt.Sprintf("%-16s", data[k].TimeBehindString[0:TimeBehindStringLen]))
	}

	return sl
}

func readData(e *ESDI, done <-chan string) {
	// dataReaderTicker := time.NewTicker(time.Second / 60)
	dataReaderTicker := time.NewTicker(time.Second / 240)
	defer dataReaderTicker.Stop()

	initialTime = time.Now()
	for {
		select {
		case <-done:
			return
		case <-dataReaderTicker.C:
			var err error

			mu.Lock()
			_, err = e.irsdk.Update(time.Millisecond * 100)
			if err != nil {
				fmt.Printf("could not update data: %v", err)
				continue
			}
			mu.Unlock()

			e.getVehicleData()
			e.fuelData()
			e.lapData()
			e.positionData()

			// Test the actual dataPacket we are sending over the wire
			copyBytes(e.dataPacket.Speed[:], SpeedLen, fmt.Sprintf("%3d", e.data.Speed))
			copyBytes(e.dataPacket.Gear[:], GearLen, fmt.Sprintf("%2d", e.data.Gear))
			copyBytes(e.dataPacket.RPM[:], RpmLen, fmt.Sprintf("%3d", e.data.RPM))

			e.dataPacket.Standings = packageStandingsLineDataPacket(e.data.Standings)

			copyBytes(e.dataPacket.LapNumber[:], LapNumberLen, fmt.Sprintf("%-3d", e.data.LapCount))
			copyBytes(e.dataPacket.DeltaToBestLap[:], DeltaToBestLapLen,
				fmt.Sprintf("%s", lapTimeDeltaRepresentation(e.data.LapDeltaFloat)))
			copyBytes(e.dataPacket.BestLapTime[:], BestLapTimeLen,
				fmt.Sprintf("%s", e.data.BestLapTime[:8]))
			copyBytes(e.dataPacket.CurrLapTime[:], CurrLapTimeLen,
				fmt.Sprintf("%s", e.data.CurrLapTime[:8]))
			copyBytes(e.dataPacket.LastLapTime[:], LastLapTimeLen,
				fmt.Sprintf("%s", e.data.LastLapTime[:8]))
			copyBytes(e.dataPacket.FuelEst[:], FuelEstLen,
				fmt.Sprintf("%.1f - %.1f", e.data.FuelPerLap, e.data.FuelLiters/e.data.FuelPerLap))

			e.dataPacket.StartMarker = 0x02
			e.dataPacket.EndMarker = 0x03

			lastMessageTime = time.Now()
		}
	}
}
