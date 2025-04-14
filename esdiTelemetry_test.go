package main

import (
	"testing"
)

func TestLapTimeDeltaRepresentationPositiveLT1(t *testing.T) {
	// Arrange
	var time float32 = 0.123
	expect := "-.1"

	// Act
	res := lapTimeDeltaRepresentation(time)

	// Test
	if res != expect {
		t.Fatalf("Expected `%s`, got `%s`", expect, res)
	}
}

func TestLapTimeDeltaRepresentationPositiveGE1(t *testing.T) {
	// Arrange
	var time float32 = 1.123
	expect := "-1.1"

	// Act
	res := lapTimeDeltaRepresentation(time)

	// Test
	if res != expect {
		t.Fatalf("Expected `%s`, got `%s`", expect, res)
	}
}

func TestLapTimeDeltaRepresentationNegativeLT1(t *testing.T) {
	// Arrange
	var time float32 = -0.123
	expect := "+.1"

	// Act
	res := lapTimeDeltaRepresentation(time)

	// Test
	if res != expect {
		t.Fatalf("Expected `%s`, got `%s`", expect, res)
	}
}

func TestLapTimeDeltaRepresentationNegativeGE1(t *testing.T) {
	// Arrange
	var time float32 = -1.123
	expect := "+1.1"

	// Act
	res := lapTimeDeltaRepresentation(time)

	// Test
	if res != expect {
		t.Fatalf("Expected `%s`, got `%s`", expect, res)
	}
}

func TestLapTimeDeltaRepresentationGreaterThan100(t *testing.T) {
	// Arrange
	var time float32 = -102.123
	expect := "+99.9"

	// Act
	res := lapTimeDeltaRepresentation(time)

	// Test
	if res != expect {
		t.Fatalf("Expected `%s`, got `%s`", expect, res)
	}
}

func TestLapTimeDeltaRepresentationGreaterNew(t *testing.T) {
	// Arrange
	var time float32 = -0.012
	expect := "+.0"

	// Act
	res := lapTimeDeltaRepresentation(time)

	// Test
	if res != expect {
		t.Fatalf("Expected `%s`, got `%s`", expect, res)
	}
}
