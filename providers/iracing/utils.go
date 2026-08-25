package iracing

import "github.com/ESilva15/goirsdk"

func IsRunning() bool {
	sdk, err := goirsdk.Init(goirsdk.Options{
		SourceType: goirsdk.SharedMemoryFile,
	})
	if err != nil {
		return false
	}

	sdk.Close()

	return true
}
