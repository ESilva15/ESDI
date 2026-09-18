package peripheral

import "errors"

var (
	ErrInvalidData       = errors.New("invalid data")
	ErrDeviceTimedOut    = errors.New("device timed out")
	ErrFailureToPackData = errors.New("failed to pack received data")
)
