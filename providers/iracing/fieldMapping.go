package iracing

import (
	"esdi/telemetry"
)

var internalToSDKFieldNames = map[telemetry.FieldID]string{
	telemetry.Speed: "Speed",
	telemetry.Gear:  "Gear",
	telemetry.RPM:   "RPM",
	telemetry.Empty: "empty",
}
