package main

import (
	"fmt"

	"github.com/mt-inside/go-lmsensors"
	"github.com/mt-inside/go-usvc"
)

func main() {
	log := usvc.GetLogger(false)

	if err := lmsensors.Init(); err != nil {
		panic(err)
	}

	sensors, err := lmsensors.Get()
	if err != nil {
		log.Error(err, "Can't get sensor readings")
	}

	for _, chip := range sensors.Chips {
		fmt.Println(chip.ID)
		for _, reading := range chip.Sensors {
			fmt.Printf("  [%s] %s: %s%s\n", reading.SensorType, reading.Name, reading.String(), reading.Unit)
		}
	}
}
