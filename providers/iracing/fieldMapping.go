package iracing

// This file maps the data from the desktop provider data structure to iRacing

import (
	"esdi/telemetry"
)

var internalToSDKFieldNames = map[telemetry.FieldID]string{
	telemetry.Speed: "Speed",
	telemetry.Gear:  "Gear",
	telemetry.RPM:   "RPM",
	// Engine Warnings
	telemetry.PitSpeedLimiter: "irsdk_pitSpeedLimiter",
	// Ajudstements
	telemetry.BrakeBias:       "dcBrakeBias",
	telemetry.ABSSetting:      "dcABS",
	telemetry.TCSetting:       "dcTractionControl",
	telemetry.ThrottleSetting: "dcThrottleShape",
	// Tire data
	telemetry.LFtempL: "LFtempCL",
	telemetry.LFtempM: "LFtempCM",
	telemetry.LFtempR: "LFtempCR",
	telemetry.RFtempL: "RFtempCL",
	telemetry.RFtempM: "RFtempCM",
	telemetry.RFtempR: "RFtempCR",
	telemetry.LRtempL: "LRtempCL",
	telemetry.LRtempM: "LRtempCM",
	telemetry.LRtempR: "LRtempCR",
	telemetry.RRtempL: "RRtempCL",
	telemetry.RRtempM: "RRtempCM",
	telemetry.RRtempR: "RRtempCR",
	// Session Data
	telemetry.SessionTime:       "SessionTime",
	telemetry.ReplaySessionTime: "ReplaySessionTime",
	telemetry.Empty:             "empty",
}
