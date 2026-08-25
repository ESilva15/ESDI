// Package devices simply defines what a device should have
// and some utils if necessary
package devices

import "esdi/telemetry"

type Device interface {
	SendData(*telemetry.TelemetryData)
}
