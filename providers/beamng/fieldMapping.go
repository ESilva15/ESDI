package beamng

// This file maps the data from the desktop provider data structure to iRacing

// var internalToSDKFieldNames = map[telemetry.FieldID]string{
// 	telemetry.Speed:     "Speed",
// 	telemetry.Gear:      "Gear",
// 	telemetry.RPM:       "RPM",
// 	telemetry.FuelLevel: "Fuel",
// 	// Engine Data
// 	telemetry.OilPress:  "OilPressure",
// 	telemetry.OilTemp:   "OilTemp",
// 	telemetry.WaterTemp: "EngTemp",
// 	// Engine Warnings
// 	telemetry.PitSpeedLimiter: "irsdk_pitSpeedLimiter",
// 	// Electrics (dash lights and so on)
// 	telemetry.LeftIndicator:  "sssssssss",
// 	telemetry.RightIndicator: "something",
// 	telemetry.Hazards:        "something",
// 	// Ajudstements
// 	telemetry.BrakeBias:       "dcBrakeBias",
// 	telemetry.ABSSetting:      "dcABS",
// 	telemetry.TCSetting:       "dcTractionControl",
// 	telemetry.ThrottleSetting: "dcThrottleShape",
// 	// Lap Data
// 	telemetry.LapLastLapTime: "LapLastLapTime",
// 	telemetry.LapNumber:      "Lap",
// 	// Tire data
// 	telemetry.LFtempL: "LFtempCL",
// 	telemetry.LFtempM: "LFtempCM",
// 	telemetry.LFtempR: "LFtempCR",
// 	telemetry.RFtempL: "RFtempCL",
// 	telemetry.RFtempM: "RFtempCM",
// 	telemetry.RFtempR: "RFtempCR",
// 	telemetry.LRtempL: "LRtempCL",
// 	telemetry.LRtempM: "LRtempCM",
// 	telemetry.LRtempR: "LRtempCR",
// 	telemetry.RRtempL: "RRtempCL",
// 	telemetry.RRtempM: "RRtempCM",
// 	telemetry.RRtempR: "RRtempCR",
// 	// Session Data
// 	telemetry.SessionTime:       "SessionTime",
// 	telemetry.ReplaySessionTime: "ReplaySessionTime",
// 	telemetry.Empty:             "empty",
// }
