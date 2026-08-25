package services

// type CDashService struct {
// Logger *slog.Logger
// CDash *cdashdisplay.CDashDisplay
// // DevClerk *peripheral.PeripheralDeviceClerk
// // Messages chan string
// // Telemetry Channel
// // streamCancel context.CancelFunc
// // TelemCh <-chan telemetry.TelemetryData
// }

// func NewCDashService(logger *slog.Logger) *CDashService {
// 	// sharedChannel := make(chan string, 10)
// 	return &CDashService{
// 		Logger: logger,
// 		CDash:  nil,
// 		// DevClerk: peripheral.NewPeripheralDeviceClerk(),
// 		// Messages: sharedChannel,
// 	}
// }

// func (ds *CDashService) FindDevice() {
// ds.Messages <- "looking for cdash display...\n"
// ds.Logger.Info("Looking for CDashDisplay")

// cdashdisplay.SetLogger(ds.Logger.With("[device]", "cdashdisplay"))

// display, err := cdashdisplay.Discover()
// if err != nil {
// 	ds.Logger.Info("didn't find cdashdisplay")
// 	ds.Messages <- "didn't find cdash display\n"
// 	return
// }
//
// ds.CDash = display
// ds.Logger.Info("found cdashdisplay on: " + display.WT.Cfg.Name)
// ds.Messages <- "found cdashdisplay on: " + display.WT.Cfg.Name + "\n"
// }

// func (ds *CDashService) CreateWindow(
// 	win *cdashdisplay.DesktopUIWindow,
// ) (*cdashdisplay.DesktopUIWindow, error) {
// 	updatedWindow, err := ds.CDash.CreateWindow(win)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return updatedWindow, nil
// }
//
// func (ds *CDashService) LoadLayout(layoutPath string) error {
// 	return ds.CDash.LoadLayout(layoutPath)
// }
//
// func (ds *CDashService) SaveLayout(layoutPath string) error {
// 	return ds.CDash.SaveLayout(layoutPath)
// }
//
// func (ds *CDashService) UnloadLayout() error {
// 	return ds.CDash.UnloadLayout()
// }
//
// func (ds *CDashService) UpdateWindow(win *cdashdisplay.DesktopUIWindow) error {
// 	ds.Messages <- fmt.Sprintf("Updating a window:\n%+v\n", win)
// 	return ds.CDash.UpdateWindow(win)
// }
//
// func (ds *CDashService) DeleteWindow(idx int16) error {
// 	return ds.CDash.DestroyWindow(idx)
// }
//
// func (ds *CDashService) ResizeWindow(idx int16, vec *helper.Vector) error {
// 	err := ds.CDash.ResizeWindow(idx, vec)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func (ds *CDashService) MoveWindow(idx int16, vec *helper.Vector) error {
// 	return ds.CDash.MoveWindow(idx, vec)
// }

// func (ds *CDashService) SetTelemetryChannel(ch <-chan telemetry.TelemetryData) {
// 	ds.TelemCh = ch
// }

// func (cds *CDashService) StartStream() {
// 	// NOTE: i'm using this pattern a whole lot. Maybe I can create a struct to handle this
// 	var ctx context.Context
// 	ctx, cds.streamCancel = context.WithCancel(context.Background())
//
// 	go cds.transmit(ctx)
// }
//
// func (cds *CDashService) StopStream() {
// 	if cds.streamCancel == nil {
// 		return
// 	}
//
// 	cds.streamCancel()
// 	cds.streamCancel = nil
// }

// INTERNAL

// func (ds *CDashService) transmit(ctx context.Context) {
// 	var isSending atomic.Bool
//
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case data, ok := <-ds.TelemCh:
// 			if !ok {
// 				return
// 			}
//
// 			if isSending.Load() {
// 				continue
// 			}
//
// 			isSending.Store(true)
//
// 			ds.CDash.SendData(&data)
// 			isSending.Store(false)
// 		}
// 	}
// }
