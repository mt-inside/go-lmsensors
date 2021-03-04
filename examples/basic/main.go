package main

import (
	"fmt"
	"log"

	"github.com/mt-inside/golmsensors"
)

func main() {
	sensors, err := golmsensors.Get()
	if err != nil {
		log.Fatalf("Can't get sensor readings: %v", err)
	}

	for _, chip := range sensors.ChipsList {
		fmt.Println(chip.ID)
		for _, reading := range chip.ReadingsList {
			fmt.Printf("  [%s] %s: %f\n", reading.SensorType, reading.Name, reading.Value)
		}
	}
}
