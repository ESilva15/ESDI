// Package conversions will host all of our unit conversion functions
package conversions

func MsToKph(v float32) int {
	return int((3600 * v) / 1000)
}
