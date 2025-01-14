package main

import (
  "testing"
)

func TestLapTimeDeltaRepresentationPositiveLT1(t *testing.T) {
  // Arrange
  var time float32 = 0.123
  expect := "-.12"

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
  expect := "-1.12"

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
  expect := "+.12"

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
  expect := "+1.12"

  // Act
  res := lapTimeDeltaRepresentation(time)

  // Test
  if res != expect {
    t.Fatalf("Expected `%s`, got `%s`", expect, res)
  }
}
