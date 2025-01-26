package main

import (
	// "esdi/sources/beamng"
	// "fmt"
	// "log"
	// "strings"
	// "time"
)

func beamngExample(esdi ESDI) {
	// bIF, err := beamng.Init("127.0.0.1", 4444)
	// if err != nil {
	// 	log.Fatalf("Failed to create BeamNG interface: %v", err)
	// }
	//
	// esdi.Source = &bIF
	//
	// lastTime := time.Now().UnixMilli()
	// lastDataSent := time.Now().UnixMilli()
	// for {
	// 	var err error
	// 	var buffer strings.Builder
	// 	buffer.WriteString("\033[?25l\033[2J\033[H")
	//
	// 	err = esdi.Source.UpdateData()
	// 	if err != nil {
	// 		fmt.Printf("could not update data: %v", err)
	// 		continue
	// 	}
	//
	// 	curGear, err := esdi.Source.GetData("Gear")
	// 	if err != nil {
	// 		log.Fatalf("could not get field `Gear`: %v", err)
	// 	}
	//
	// 	curRPM, err := esdi.Source.GetData("RPM")
	// 	if err != nil {
	// 		log.Fatalf("could not get field `RPM`: %v", err)
	// 	}
	//
	// 	gear := int(curGear.(int8))
	// 	rpm := int(curRPM.(float32))
	//
	// 	buffer.WriteString(fmt.Sprintf("Gear: %d, RPM: %d", gear, rpm))
	//
	// 	curTime := time.Now().UnixMilli()
	// 	message := fmt.Sprintf("%d,%d\n", gear-1, rpm)
	// 	buffer.WriteString("\n" + message)
	//
	// 	messageWasSentMark := "N"
	// 	if curTime-lastDataSent > 75 {
	// 		_, err = esdi.SerialConn.Write([]byte(message))
	// 		if err != nil {
	// 			log.Printf("Unable to write data: %v", err)
	// 			break
	// 		}
	//
	// 		messageWasSentMark = "Y"
	// 		lastDataSent = curTime
	// 	}
	//
	// 	buffer.WriteString(" -> " + messageWasSentMark)
	// 	if curTime-lastTime > 100 {
	// 		fmt.Print(buffer.String())
	// 		lastTime = curTime
	// 	}
	// }
}
