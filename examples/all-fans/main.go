package main

import (
	"fmt"
	"os"
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
		sensors, err := lmsensors.Get()
		if err != nil {
			log.Error(err, "Can't get sensor readings")
			os.Exit(1)
		}

		var ss []string
		for _, chip := range sensors.Chips {
			for _, reading := range chip.Sensors {
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
