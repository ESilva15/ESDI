package iracing

import (
	"github.com/ESilva15/goirsdk"
)

func IsRunning() bool {
	sdk, err := goirsdk.Init(goirsdk.Options{
		SourceType: goirsdk.SharedMemoryFile,
	})
	defer sdk.Close()
	if err != nil {
		return false
	}

	// if sdk.SessionStateInvalid() {
	// 	slog.Debug("state is invalid")
	// 	sessionState, ok := sdk.Vars.Vars["SessionState"]
	// 	if !ok {
	// 		slog.Debug("SessionState not ok")
	// 		return false
	// 	}
	//
	// 	state, ok := sessionState.Value.(int)
	// 	if !ok {
	// 		slog.Debug(fmt.Sprintf("can't typecast: %+v", sessionState))
	// 		return false
	// 	}
	//
	// 	slog.Debug(fmt.Sprintf("%d", state))
	// 	return false
	// }

	return true
}

func (i *IRacing) Stalled() bool {
	return i.Stalled()
}
