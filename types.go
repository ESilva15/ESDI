package main

const (
	DriverNameLen       = 24
	TimeBehindStringLen = 16
	FuelTankLen         = 15 // 101.6 / 123.4L
)

type GameSource interface {
	GetData(string) (interface{}, error)
	UpdateData() error
	GetSessionInfo() (interface{}, error)
}

type StandingsLine struct {
	CarIdx           int32
	LapPct           float32
	Lap              int32
	DriverName       [DriverNameLen]byte
	EstTime          float32
	TimeBehind       float32
	TimeBehindString [TimeBehindStringLen]byte
}

type SimulationData struct {
	Speed           int32
	Gear            int32
	RPM             int32
	LapCount        int32
	LapDistPct      float32
	LapTime         [16]byte // Current lap time
	LapDelta        [16]byte // Delta to selected reference lap
	BestLapTime     [16]byte // Best lap in session
	LastLapTime     [16]byte // Last lap time
	FuelUsageCurLap float32
	FuelPerLap      float32
	FuelPct         float32
	FuelLiters      float32
	FuelTotal       float32 // This will be calculated and passed in Liters
	Position        int32
	Standings       [5]StandingsLine
}

type StandingsLineDataPacket struct {
	Lap              [4]byte
	DriverName       [DriverNameLen]byte
	TimeBehindString [TimeBehindStringLen]byte
}

type DataPacket struct {
	Speed       int32
	Gear        int32
	RPM         int32
	LapCount    int32
	LapTime     [10]byte // Current lap time
	LapDelta    [10]byte // Delta to selected reference lap
	BestLapTime [10]byte // Best lap in session
	LastLapTime [10]byte // Last lap time
	FuelTank    [FuelTankLen]byte
	FuelPerLap  [8]byte
	FuelPct     [8]byte
	FuelLiters  [8]byte
	FuelTotal   [8]byte // This will be calculated and passed in Liters
	Position    int32
	Standings   [5]StandingsLineDataPacket
}
