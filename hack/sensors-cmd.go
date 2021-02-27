/* NOTE
* I never finished this, but it just needs:
* - Sensor type (in, fan, temp, etc) - will have to be extracted from the name prefixes
* - Temperature sensor type (tempX_type) - can be interpreted as https://github.com/lm-sensors/lm-sensors/blob/42f240d2a457834bcbdf4dc8b57237f97b5f5854/prog/sensors/chips.c#L407
* - Units: are implicit. In: V, Fan: 1/min, Temp: C. I'm 99% sure that temp is always degC; `sensors` will print F, but that's a feature of that UI; libsensors/hwmon deals only in C (https://github.com/lm-sensors/lm-sensors/blob/42f240d2a457834bcbdf4dc8b57237f97b5f5854/prog/sensors/chips.c#L388)
 */

package sensors

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Sensors struct {
	Chips map[string]Chip
}

type Chip struct {
	Id      string
	Type    string
	Bus     string
	Address string
	Adapter string

	Readings map[string]Reading
}

type Reading struct {
	Name  string
	Type  SensorType
	Value float64
	Alarm bool
}

//go:generate stringer -type=SensorType
type SensorType int

const (
	Voltage SensorType = iota
	Temperature
	Tacho
	Alarm
)

func Get() (Sensors, error) {
	jsonBytes, err := exec.Command("sensors", "-j").Output()
	if err != nil {
		return Sensors{}, fmt.Errorf("Failed to get sensor readings: %w", err)
	}

	var j interface{}
	err = json.Unmarshal(jsonBytes, &j)
	if err != nil {
		return Sensors{}, fmt.Errorf("Failed to unmarshall sensor json: %w", err)
	}

	sensors := Sensors{Chips: make(map[string]Chip)}

	chipsMap := j.(map[string]interface{})
	for k, v := range chipsMap {
		chip := Chip{Id: k, Readings: make(map[string]Reading)}
		ords := strings.Split(k, "-")
		chip.Type = ords[0]
		chip.Bus = ords[1]
		chip.Address = ords[2]

		readingsMap := v.(map[string]interface{})
		chip.Adapter = readingsMap["Adapter"].(string)
		delete(readingsMap, "Adapter")
		for k, v := range readingsMap {
			reading := Reading{Name: k}

			valuesMap := v.(map[string]interface{})
			for k, v := range valuesMap {
				if strings.HasSuffix(k, "_input") {
					reading.Value = v.(float64)
				} else if strings.HasSuffix(k, "_alarm") {
					reading.Alarm = v.(float64) == 1.0
				}
				// TODO detect type (better than name prefix?)
			}

			chip.Readings[k] = reading
		}

		sensors.Chips[k] = chip
	}

	return sensors, nil
}

func (s Sensors) String() string {
	out, _ := json.Marshal(s)
	return string(out)
}
