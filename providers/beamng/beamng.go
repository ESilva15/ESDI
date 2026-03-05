package beamng

import (
	"fmt"

	"github.com/ESilva15/gobngsdk"
)

// This will implement the GameSink interface from the main package
type BeamNG struct {
	SDK *bngsdk.BNGSDK
}

const (
	NAME = "BeamNG.Drive"
)

func Init(ip string, port int) (BeamNG, error) {
	var err error

	sdk, err := bngsdk.Init(ip, port)
	if err != nil {
		return BeamNG{}, err
	}

	return BeamNG{SDK: &sdk}, nil
}

// GetData will retrieve a given field by its name from the OutGauge data
func (b *BeamNG) GetData(fieldName string) (interface{}, error) {
	if val, ok := b.SDK.DataDict[fieldName]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("key `%s` doesn't exist", fieldName)
}

func (b *BeamNG) UpdateData() error {
	return b.SDK.ReadData()
}

func (b *BeamNG) GetSessionInfo() (interface{}, error) {
  return nil, fmt.Errorf("Not implemented")
}
