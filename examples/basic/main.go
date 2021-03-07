package main

import (
	"fmt"

	"github.com/mt-inside/go-lmsensors"
	"github.com/mt-inside/go-usvc"
)

func main() {
	log := usvc.GetLogger(false)

	sensors, err := lmsensors.Get(true)
	if err != nil {
		log.Error(err, "Can't get sensor readings")
	}

	for _, chip := range sensors.ChipsList {
		fmt.Println(chip.ID)
		for _, reading := range chip.SensorsList {
			fmt.Printf("  [%s] %s: %s%s\n", reading.SensorType, reading.Name, reading.Value, reading.Unit)
		}
	}
}
