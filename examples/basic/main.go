package main

import (
	"fmt"

	"github.com/mt-inside/go-usvc"

	"github.com/mt-inside/go-lmsensors"
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
		fmt.Println(chip.String())
		for _, reading := range chip.Sensors {
			fmt.Println("  " + reading.String())
		}
	}
}
