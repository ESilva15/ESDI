package iracing

import (
	"esilva.org.localhost/bngsdk"
)

// This will implement the GameSink interface from the main package
type BeamNG struct {
	SDK *bngsdk.BNGSDK
}

const (
	NAME = "iRacing"
)

func Init(ip string, port int) (BeamNG, error) {
	var err error

	sdk, err := bngsdk.Init(ip, port)
	if err != nil {
		return BeamNG{}, err
	}

	return BeamNG{SDK: &sdk}, nil
}

func (b *BeamNG) GetData(fieldName string, out interface{}) error {
	return nil
}

func (b *BeamNG) UpdateData() error {
	return b.SDK.ReadData()
}
