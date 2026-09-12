package iracing

import (
	"log/slog"
	"time"

	"github.com/ESilva15/goirsdk"
	eventutils "github.com/ESilva15/goirsdk/eventutils"
)

func IsRunning() bool {
	irUtils, err := eventutils.Init()
	if err != nil {
		// we need error validation or something here
		return false
	}
	defer irUtils.Close()

	if err := irUtils.OpenEvent(goirsdk.IRSDK_DATAVALIDEVENTNAME); err != nil {
		// we need error validation or something here
		return false
	}

	// We now check for some consecutive data events
	for range 3 {
		if !irUtils.CheckValidDataEvent(1 * time.Second) {
			slog.Debug("timed out waiting for DataValidEvent")
			return false
		}
	}

	return true
}
