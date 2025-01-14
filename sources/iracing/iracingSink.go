package iracing

import (
	"fmt"
	"time"

	"github.com/ESilva15/goirsdk"
)

// This will implement the GameSink interface from the main package
type IRacing struct {
	SDK *goirsdk.IBT
}

const (
	NAME = "iRacing"
)

func Init(f goirsdk.Reader, telemOut string, yamlOut string) (IRacing, error) {
	var err error

	sdk, err := goirsdk.Init(f, telemOut, yamlOut)
	if err != nil {
		return IRacing{}, err
	}

	return IRacing{SDK: sdk}, nil
}

func (i *IRacing) GetData(fieldName string) (interface{}, error) {
	if val, ok := i.SDK.Vars.Vars[fieldName]; ok {
		return val.Value, nil
	}

	return nil, fmt.Errorf("key `%s` doesn't exist", fieldName)
}

func (i *IRacing) UpdateData() error {
	_, err := i.SDK.Update(100 * time.Millisecond)
	return err
}

func (i *IRacing) GetSessionInfo() (interface{}, error) {
  return i.SDK.SessionInfo, nil
}
