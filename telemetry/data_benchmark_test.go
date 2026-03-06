package telemetry

import (
	"testing"
)

// Old Way
type TelemetryFieldAny struct {
	Value any
}

// New Way
type TelemetryFieldUint struct {
	Raw uint64
}

func BenchmarkInterfaceBoxing(b *testing.B) {
	var field TelemetryFieldAny
	b.ReportAllocs() // Tracks memory allocations
	for i := 0; i < b.N; i++ {
		// Every iteration, we put a uint16 into 'any'
		// This forces a heap allocation in many Go versions
		field.Value = uint16(i % 65535)
	}
	_ = field
}

func BenchmarkUint64Bucket(b *testing.B) {
	var field TelemetryFieldUint
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Casting to uint64 is a direct CPU register move
		field.Raw = uint64(uint16(i % 65535))
	}
	_ = field
}
