package iracing

import (
	"log"
	"log/slog"
	"time"

	"github.com/ESilva15/goirsdk"
	eventutils "github.com/ESilva15/goirsdk/eventutils"
)

func IsRunning() bool {
	irUtils, err := eventutils.Init()
	if err != nil {
		log.Fatalf("Failed to initialize mmaputils: %v", err)
	}
	defer irUtils.Close()

	// 2. Connect to the OS event (Win32 Event on Windows, POSIX Semaphore on Linux)
	if err := irUtils.OpenEvent(goirsdk.IRSDK_DATAVALIDEVENTNAME); err != nil {
		log.Fatalf("Failed to open event: %v", err)
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

func (i *IRacing) Stalled() bool {
	return i.Stalled()
}
