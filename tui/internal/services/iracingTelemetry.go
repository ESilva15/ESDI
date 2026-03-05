package services

// NOTE: I have to think about how to implement this, for now I'm bruteforcing
// iracing support only. But I should have a generic service that can gather
// telemetry from multiples sims and just publish it in an internal format

// Car data lengths
const (
	SpeedLen     = 5
	GearLen      = 3
	RpmLen       = 6
	BrakeBiasLen = 6
)

type StreamState int

const (
	StreamStatePaused  StreamState = 0
	StreamStateRunning StreamState = 1
	StreamStateOff     StreamState = 2
)

// type IRacingService struct {
// 	Message chan string
// 	// Timers
// 	LastMessageTime time.Time
// 	InitialTime     time.Time
// 	LastTime        time.Time
// 	ticker          *time.Ticker
// 	// Data vessels
// 	data     *esdi.SimulationData
// 	dataView *esdi.DataPacket
// 	// Data access control
// 	Mut sync.Mutex
// 	// Source
// 	Irsdk *goirsdk.IBT
// 	// Stream control
// 	isRunning    bool
// 	UIStream     chan string
// 	DataStream   chan *esdi.DataPacket
// 	StreamCancel context.CancelFunc
// }

// func NewIRacingService(msg chan string) *IRacingService {
// 	// Open the telemetry file
// 	file, err := os.Open("/home/esilva/Desktop/projetos/simracing_peripherals/testTelemetry/supercars_indianapolis.ibt")
// 	if err != nil {
// 		log.Fatalf("Failed to open IBT file: %v", err)
// 	}
//
// 	irsdk, err := goirsdk.Init(file, "./out.ibt", "./out.yaml")
// 	if err != nil {
// 		log.Fatalf("Failed to load iRacing data")
// 	}
//
// 	return &IRacingService{
// 		Message:    msg,
// 		ticker:     time.NewTicker(time.Second / 60),
// 		Irsdk:      irsdk,
// 		UIStream:   make(chan string, 10),
// 		DataStream: make(chan *esdi.DataPacket),
// 		data:       &esdi.SimulationData{},
// 		dataView:   &esdi.DataPacket{},
// 		isRunning:  false,
// 	}
// }

// func (irs *IRacingService) GetStream() <-chan string {
// 	return irs.UIStream
// }
//
// func (irs *IRacingService) StartStream() <-chan *esdi.DataPacket {
// 	if irs.isRunning {
// 		return nil
// 	}
//
// 	var ctx context.Context
// 	ctx, irs.StreamCancel = context.WithCancel(context.Background())
//
// 	irs.startStream(ctx)
// 	irs.isRunning = true
//
// 	return irs.DataStream
// }

// func (irs *IRacingService) StopStream() {
// 	if irs.StreamCancel != nil {
// 		irs.StreamCancel()
// 	}
// }
