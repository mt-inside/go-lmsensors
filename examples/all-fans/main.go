package main

import (
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mt-inside/go-usvc"

	"github.com/mt-inside/go-lmsensors"
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

		var chipIDs []string
		for _, chip := range system.Chips {
			chipIDs = append(chipIDs, chip.ID)
		}
		sort.Strings(chipIDs)

		for _, chipID := range chipIDs {
			chip := system.Chips[chipID]

			var sensorNames []string
			for _, sensor := range chip.Sensors {
				sensorNames = append(sensorNames, sensor.GetName())
			}
			sort.Strings(sensorNames)

			for _, s := range sensorNames {
				reading := chip.Sensors[s]

				if fan, ok := reading.(*lmsensors.FanSensor); ok && fan.Value != 0.0 {
					ss = append(ss, fan.String())
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
