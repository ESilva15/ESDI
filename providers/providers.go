// Package providers
package providers

import (
	"esdi/providers/beamng"
	"esdi/providers/iracing"
)

// Make this be some kind of struct where we can access a function that returns
// the selected provider by its name
type Provider struct {
	Name string
}

var Providers = map[string]Provider{
	beamng.NAME: {
		Name: beamng.NAME,
	},
	iracing.NAME: {
		Name: iracing.NAME,
	},
}
