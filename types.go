package main

// DataPacket Lens
const (
	SpeedLen          = 5
	GearLen           = 3
	RpmLen            = 6
	LapNumberLen      = 5
	DeltaToBestLapLen = 6
	BestLapTimeLen    = 10
	CurrLapTimeLen    = 10
	LastLapTimeLen    = 10
	FuelTankLen       = 15 // 101.6 / 123.4L
	FuelEstLen        = 15
)

// StandingsLineDataPacket Lens
const (
	LapStringLen        = 4
	DriverNameLen       = 24
	TimeBehindStringLen = 8
)

type GameSource interface {
	GetData(string) (any, error)
	UpdateData() error
	GetSessionInfo() (any, error)
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
	ReadError       error
	Recv            int8
	Speed           int32
	Gear            int32
	RPM             int32
	LapCount        int32
	LapDistPct      float32
	CurrLapTime     [16]byte // Current lap time
	LapDeltaFloat   float32
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
	Lap              [LapStringLen]byte        `binary:"little"`
	DriverName       [DriverNameLen]byte       `binary:"little"`
	TimeBehindString [TimeBehindStringLen]byte `binary:"little"`
}

type DataPacket struct {
	StartMarker    uint8                      `binary:"little"`
	Speed          [SpeedLen]byte             `binary:"little"`
	Gear           [GearLen]byte              `binary:"little"`
	RPM            [RpmLen]byte               `binary:"little"`
	LapNumber      [LapNumberLen]byte         `binary:"little"`
	DeltaToBestLap [DeltaToBestLapLen]byte    `binary:"little"`
	BestLapTime    [BestLapTimeLen]byte       `binary:"little"`
	CurrLapTime    [CurrLapTimeLen]byte       `binary:"little"`
	LastLapTime    [LastLapTimeLen]byte       `binary:"little"`
	FuelEst        [FuelEstLen]byte           `binary:"little"`
	Standings      [5]StandingsLineDataPacket `binary:"little"`
	EndMarker      uint8                      `binary:"little"`
}
