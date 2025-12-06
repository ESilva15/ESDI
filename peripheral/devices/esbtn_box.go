package devices

import helper "esdi/helpers"

var ESBtnBox = Device{
	ID:   ESBtnBoxDevID,
	Name: helper.B32("ESLabs Button Box"),
	API:  map[string]DeviceCMD{},
}
