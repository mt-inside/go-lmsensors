package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mt-inside/go-lmsensors"
	"github.com/mt-inside/go-usvc"
)

func main() {
	log := usvc.GetLogger(false)
	signalCh := usvc.InstallSignalHandlers(log)

	if err := lmsensors.Init(); err != nil {
		panic(err)
	}

	for {
		system, err := lmsensors.Get()
		if err != nil {
			log.Error(err, "Can't get sensor readings")
			os.Exit(1)
		}

		var ss []string

		// This shows how to sort the bus and device maps to get a stable ordering over them.
		// In this example programme we could actually have just sorted the output
		for _, chipId := range sortedChipIds(system.Chips) {
			chip := system.Chips[chipId]

			for _, s := range sortedSensorIds(chip.Sensors) {
				reading := chip.Sensors[s]

				if reading.SensorType == lmsensors.Fan && reading.Value != 0.0 {
					ss = append(ss, fmt.Sprintf("%s: %s", reading.Name, reading.Rendered))
				}
			}
		}

		usvc.PrintUpdateLn(strings.Join(ss, "\t"))

		select {
		case <-signalCh:
			return
		case <-time.After(1 * time.Second):
			continue
		}
	}
}

func sortedChipIds(m map[string]*lmsensors.Chip) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func sortedSensorIds(m map[string]*lmsensors.Sensor) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
