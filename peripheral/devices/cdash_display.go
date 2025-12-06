package devices

import helper "esdi/helpers"

const (
	newWindowCMDID = iota
	destroyWindowCMDID
)

var CDashDisplay = Device{
	ID:   CDashDisplayDevID,
	Name: helper.B32("ESLabs CDashDisplay"),
	API: map[string]DeviceCMD{
		"new-window": {
			Identifier: newWindowCMDID,
			Name:       "new-window",
			Desc:       "Creates a new window - pass the window name and dimensions",
			Fn:         createWindow,
		},
		"destroy-window": {
			Identifier: destroyWindowCMDID,
			Name:       "destroy-window",
			Desc:       "Destroys a window by its ID",
			Fn:         destroyWindow,
		},
	},
}

func createWindow(data []byte) {
	// Parse the command
}

func destroyWindow(data []byte) {
	// Parse the command
}
