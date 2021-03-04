package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mt-inside/golmsensors"
	"github.com/mt-inside/logging"
)

func main() {
	log := logging.GetLogger(false)
	signalCh := logging.InstallSignalHandlers(log)

forever:
	for {
		sensors, err := golmsensors.Get()
		if err != nil {
			log.Error(err, "Can't get sensor readings")
			os.Exit(1)
		}

		for _, chip := range sensors.ChipsList {
			for _, reading := range chip.ReadingsList {
				if reading.SensorType == golmsensors.Fan && reading.Value != 0 {
					fmt.Printf("%s: %d    ", reading.Name, int(reading.Value))
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
