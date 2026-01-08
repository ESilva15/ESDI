// Package esdi
package esdi

import "esdi/peripheral"

type ESDI struct {
	PeripheralClerk *peripheral.PeripheralDeviceClerk
}

func NewESDI() *ESDI {
	return &ESDI{
		PeripheralClerk: peripheral.NewPeripheralDeviceClerk(),
	}
}
