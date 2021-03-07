package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mt-inside/go-lmsensors"
	"github.com/mt-inside/go-usvc"
)

func main() {
	log := usvc.GetLogger(false)
	signalCh := usvc.InstallSignalHandlers(log)

forever:
	for {
		sensors, err := lmsensors.Get(true)
		if err != nil {
			log.Error(err, "Can't get sensor readings")
			os.Exit(1)
		}

		for _, chip := range sensors.ChipsList {
			for _, reading := range chip.SensorsList {
				if reading.SensorType == lmsensors.Fan && reading.Value != "0" {
					fmt.Printf("%s: %s    ", reading.Name, reading.Value)
				}
			}
		}
		fmt.Printf("\r")

		select {
		case <-signalCh:
			break forever
		case <-time.After(1 * time.Second):
			continue
		}
	}
}
