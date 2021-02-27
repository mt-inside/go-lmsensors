/*
 * golmsensors
 *
 * Copyright (c) 2021 Matt Turner.
 */

package golmsensors

// #include <sensors/sensors.h>
// #include <sensors/error.h>
// #cgo LDFLAGS: -lsensors
import "C"

import (
	"fmt"
	"math"
	"strings"
)

//go:generate stringer -type=SensorType
type SensorType int

// https://github.com/lm-sensors/lm-sensors/blob/42f240d2a457834bcbdf4dc8b57237f97b5f5854/lib/sensors.h#L138
const (
	In       SensorType = 0x00
	Fan      SensorType = 0x01
	Temp     SensorType = 0x02
	Power    SensorType = 0x03
	Energy   SensorType = 0x04
	Curr     SensorType = 0x05
	Humidity SensorType = 0x06

	VID        SensorType = 0x10
	Intrustion SensorType = 0x11

	BeepEnable SensorType = 0x18

	Unhandled SensorType = math.MaxInt32
)

//go:generate stringer -type=TempType
type TempType int

// Not defined in a library header, but: https://github.com/lm-sensors/lm-sensors/blob/42f240d2a457834bcbdf4dc8b57237f97b5f5854/prog/sensors/chips.c#L407
const (
	Disabled     TempType = 0
	CPUDiode     TempType = 1
	Transistor   TempType = 2
	ThermalDiode TempType = 3
	Thermistor   TempType = 4
	AMDSI        TempType = 5 // ??
	IntelPECI    TempType = 6 // Platform Environment Control Interface
)

type Sensors struct {
	Chips map[string]*Chip
}

type Chip struct {
	ID      string
	Type    string
	Bus     string
	Address string
	Adapter string

	Readings map[string]Reading
}

type Reading struct {
	Name       string
	SensorType SensorType
	Value      float64
	Alarm      bool

	// TODO: make a separate type with ^^ embedded, plus this, plus an interface over them.
	TempType TempType
}

func getValue(chip *C.sensors_chip_name, sf *C.struct_sensors_subfeature) (C.double, error) {
	var val C.double
	var err error

	cerr := C.sensors_get_value(chip, sf.number, &val)
	if cerr != 0 {
		err = fmt.Errorf("Can't read sensor value: chip=%v, subfeature=%v, error=%d", chip, sf, cerr)
	}

	return val, err
}

func Get() (*Sensors, error) {
	cerr := C.sensors_init(nil)
	if cerr != 0 {
		panic("Can't configure libsensors")
	}

	sensors := &Sensors{Chips: make(map[string]*Chip)}

	var chipno C.int = 0
	for {
		cchip := C.sensors_get_detected_chips(nil, &chipno)
		if cchip == nil {
			break
		}

		chipNameBuf := strings.Repeat(" ", 200)
		cchipNameBuf := C.CString(chipNameBuf)
		C.sensors_snprintf_chip_name(cchipNameBuf, C.ulong(len(chipNameBuf)), cchip)
		chipName := C.GoString(cchipNameBuf)

		adaptor := C.GoString(C.sensors_get_adapter_name(&cchip.bus))

		chip := &Chip{ID: chipName, Adapter: adaptor, Readings: make(map[string]Reading)}
		ords := strings.Split(chipName, "-")
		chip.Type = ords[0]
		chip.Bus = ords[1]
		chip.Address = ords[2]

		i := C.int(0)
		for {
			feature := C.sensors_get_features(cchip, &i)
			if feature == nil {
				break
			}
			sensorType := SensorType(feature._type)

			clabel := C.sensors_get_label(cchip, feature)
			if clabel == nil {
				continue
			}
			label := C.GoString(clabel)

			reading := Reading{Name: label, SensorType: sensorType}

			switch sensorType {
			case Temp:
				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_TEMP_INPUT)
				if sf != nil {
					cvalue, _ := getValue(cchip, sf)
					reading.Value = float64(cvalue)
				}

				sf = C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_TEMP_TYPE)
				if sf != nil {
					cvalue, _ := getValue(cchip, sf)
					reading.TempType = TempType(C.int(cvalue))
				}

				//TODO
				reading.Alarm = false

			case In:
				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_IN_INPUT)
				if sf != nil {
					cvalue, _ := getValue(cchip, sf)
					reading.Value = float64(cvalue)
				}

				//TODO
				reading.Alarm = false

			case Fan:
				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_FAN_INPUT)
				if sf != nil {
					cvalue, _ := getValue(cchip, sf)
					reading.Value = float64(cvalue)
				}

				//TODO
				reading.Alarm = false
			}
			chip.Readings[label] = reading
		}
		sensors.Chips[chipName] = chip
	}

	return sensors, nil
}
