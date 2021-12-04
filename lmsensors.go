/*
 * go-lmsensors
 *
 * Copyright (c) 2021 Matt Turner.
 */

package lmsensors

// #include <sensors/sensors.h>
// #include <sensors/error.h>
// #cgo LDFLAGS: -lsensors
import "C"

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

//go:generate stringer -type=SensorType
// SensorType is the type of sensor (eg Temperature or Fan RPM)
type SensorType int

// https://github.com/lm-sensors/lm-sensors/blob/42f240d2a457834bcbdf4dc8b57237f97b5f5854/lib/sensors.h#L138
const (
	Voltage     SensorType = 0x00
	Fan         SensorType = 0x01
	Temperature SensorType = 0x02
	Power       SensorType = 0x03
	Energy      SensorType = 0x04
	Current     SensorType = 0x05
	Humidity    SensorType = 0x06

	VID       SensorType = 0x10
	Intrusion SensorType = 0x11

	BeepEnable SensorType = 0x18

	Unhandled SensorType = math.MaxInt32
)

//go:generate stringer -type=TempType
// TempType is the type of temperature sensor (eg Thermistor or Diode)
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

// System contains all the chips, and all their sensors, in the system
type System struct {
	Chips          map[string]*Chip
	ChipKeysSorted []string
}

// Chip represents a hardware monitoring chip, which has one or more sensors attached, possibly of different types.
type Chip struct {
	ID      string
	Type    string
	Bus     string
	Address string
	Adapter string

	Sensors          map[string]*Sensor
	SensorKeysSorted []string
}

// Sensor represents one monitoring sensor, its type (temperature, voltage, etc), and its reading.
type Sensor struct {
	Name       string
	SensorType SensorType
	Unit       string
	Value      float64
	Rendered   string
	Alarm      bool

	// TODO: make a separate type with ^^ embedded, plus this, plus an interface over them.
	TempType TempType
}

// Init initialises the underlying lmsensors library, eg loading its database of sensor names and curves.
func Init() error {
	cerr := C.sensors_init(nil)
	if cerr != 0 {
		return errors.New("Can't configure libsensors")
	}
	return nil
}

func getValue(chip *C.sensors_chip_name, sf *C.struct_sensors_subfeature) (float64, error) {
	var val C.double

	cerr := C.sensors_get_value(chip, sf.number, &val)
	if cerr != 0 {
		return 0.0, fmt.Errorf("Can't read sensor value: chip=%v, subfeature=%v, error=%d", chip, sf, cerr)
	}

	return float64(val), nil
}

// Get fetches all the chips, all their sensors, and all their values.
func Get() (*System, error) {
	sensors := &System{Chips: map[string]*Chip{}}

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

		chip := &Chip{ID: chipName, Adapter: adaptor, Sensors: map[string]*Sensor{}}
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

			reading := &Sensor{Name: label, SensorType: sensorType}

			switch sensorType {
			case Temperature:
				reading.Unit = "°C"

				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_TEMP_INPUT)
				if sf != nil {
					value, _ := getValue(cchip, sf)
					reading.Value = value
					reading.Rendered = strconv.FormatFloat(value, 'f', -1, 64)
				}

				sf = C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_TEMP_TYPE)
				if sf != nil {
					value, _ := getValue(cchip, sf)
					reading.TempType = TempType(int(value))
				}

				//TODO
				reading.Alarm = false

			case Voltage:
				reading.Unit = "V"

				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_IN_INPUT)
				if sf != nil {
					value, _ := getValue(cchip, sf)
					reading.Rendered = strconv.FormatFloat(value, 'f', 2, 64)
					reading.Value = value
				}

				//TODO
				reading.Alarm = false

			case Fan:
				reading.Unit = "/min"

				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_FAN_INPUT)
				if sf != nil {
					value, _ := getValue(cchip, sf)
					reading.Rendered = strconv.FormatFloat(value, 'f', 0, 64)
					reading.Value = value
				}

				//TODO
				reading.Alarm = false
			}

			chip.SensorKeysSorted = append(chip.SensorKeysSorted, reading.Name)
			chip.Sensors[reading.Name] = reading
		}
		sort.Strings(chip.SensorKeysSorted)

		sensors.ChipKeysSorted = append(sensors.ChipKeysSorted, chip.ID)
		sensors.Chips[chip.ID] = chip
	}
	sort.Strings(sensors.ChipKeysSorted)

	return sensors, nil
}
