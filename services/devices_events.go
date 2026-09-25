package services

func (ds *DeviceService) OnStoppedMidStream() {
	// Need to make every device be unconfigured
	ds.StopStream()
}
