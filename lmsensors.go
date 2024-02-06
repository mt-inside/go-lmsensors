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
	"fmt"
	"math"
	"strconv"
	"strings"
)

// LmSensorType is the type of sensor (eg Temperature or Fan RPM)
//
//go:generate stringer -type=LmSensorType
type LmSensorType int

// https://github.com/lm-sensors/lm-sensors/blob/42f240d2a457834bcbdf4dc8b57237f97b5f5854/lib/sensors.h#L138
const (
	Voltage     LmSensorType = 0x00
	Fan         LmSensorType = 0x01
	Temperature LmSensorType = 0x02
	Power       LmSensorType = 0x03
	Energy      LmSensorType = 0x04
	Current     LmSensorType = 0x05
	Humidity    LmSensorType = 0x06

	VID       LmSensorType = 0x10
	Intrusion LmSensorType = 0x11

	BeepEnable LmSensorType = 0x18

	Unhandled LmSensorType = math.MaxInt32
)

// System contains all the chips, and all their sensors, in the system
type System struct {
	Chips map[string]*Chip
}

// Chip represents a hardware monitoring chip, which has one or more sensors attached, possibly of different types.
type Chip struct {
	ID      string
	Type    string
	Bus     string
	Address string
	Adapter string

	Sensors map[string]Sensor
}

func (c *Chip) String() string {
	return fmt.Sprintf("%s at %s:%s", c.Type, c.Bus, c.Address)
}

// Sensor represents one monitoring sensor, its type (temperature, voltage, etc), and its reading.
type Sensor interface {
	fmt.Stringer

	GetName() string
	Rendered() string
	Unit() string
	Alarm() bool
}

type baseSensor struct {
	Name  string
	Value float64
}

func (s *baseSensor) GetName() string {
	return s.Name
}

// LmTempType is the type of temperature sensor (eg Thermistor or Diode)
//
//go:generate stringer -type=LmTempType
type LmTempType int

// Not defined in a library header, but: https://github.com/lm-sensors/lm-sensors/blob/42f240d2a457834bcbdf4dc8b57237f97b5f5854/prog/sensors/chips.c#L407
const (
	Disabled     LmTempType = 0
	CPUDiode     LmTempType = 1
	Transistor   LmTempType = 2
	ThermalDiode LmTempType = 3
	Thermistor   LmTempType = 4
	AMDSI        LmTempType = 5 // ??
	IntelPECI    LmTempType = 6 // Platform Environment Control Interface

	Unknown LmTempType = math.MaxInt32
)

type TempSensor struct {
	baseSensor

	TempType LmTempType
}

func (s *TempSensor) Rendered() string {
	return strconv.FormatFloat(s.Value, 'f', 0, 64)
}

func (s *TempSensor) Unit() string {
	return "°C"
}

func (s *TempSensor) Alarm() bool {
	return false
}

func (s *TempSensor) String() string {
	var ret strings.Builder
	fmt.Fprintf(&ret, "%s: %s%s", s.Name, s.Rendered(), s.Unit())
	if s.TempType != Unknown {
		fmt.Fprintf(&ret, " (%s)", s.TempType)
	}
	return ret.String()
}

type VoltageSensor struct {
	baseSensor
}

func (s *VoltageSensor) Rendered() string {
	return strconv.FormatFloat(s.Value, 'f', 2, 64)
}

func (s *VoltageSensor) Unit() string {
	return "V"
}

func (s *VoltageSensor) Alarm() bool {
	return false
}

func (s *VoltageSensor) String() string {
	return fmt.Sprintf("%s: %s%s", s.Name, s.Rendered(), s.Unit())
}

type FanSensor struct {
	baseSensor
}

func (s *FanSensor) Rendered() string {
	return strconv.FormatFloat(s.Value, 'f', 0, 64)
}

func (s *FanSensor) Unit() string {
	return "min⁻¹"
}

func (s *FanSensor) Alarm() bool {
	return false
}

func (s *FanSensor) String() string {
	return fmt.Sprintf("%s: %s%s", s.Name, s.Rendered(), s.Unit())
}

type UnimplementedSensor struct {
	baseSensor

	sensorType LmSensorType
}

func (s *UnimplementedSensor) Rendered() string {
	return strconv.FormatFloat(s.Value, 'f', 2, 64)
}

func (s *UnimplementedSensor) Unit() string {
	return "TODO"
}

func (s *UnimplementedSensor) Alarm() bool {
	return false
}

func (s *UnimplementedSensor) String() string {
	return fmt.Sprintf("[UNIMPLEMENTED SENSOR TYPE: %s; name: %s]", s.sensorType, s.Name)
}

// Init initialises the underlying lmsensors library, eg loading its database of sensor names and curves.
func Init() error {
	cerr := C.sensors_init(nil)
	if cerr != 0 {
		return fmt.Errorf("can't configure libsensors: sensors_init() return code: %d", cerr)
	}

	return nil
}

func getValue(chip *C.sensors_chip_name, sf *C.struct_sensors_subfeature) (float64, error) {
	var val C.double

	cerr := C.sensors_get_value(chip, sf.number, &val)
	if cerr != 0 {
		return 0.0, fmt.Errorf("can't read sensor value: chip=%v, subfeature=%v, error=%d", chip, sf, cerr)
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

		chipNameBuf := strings.Repeat(" ", 256)
		cchipNameBuf := C.CString(chipNameBuf)
		C.sensors_snprintf_chip_name(cchipNameBuf, C.ulong(len(chipNameBuf)), cchip)
		chipName := C.GoString(cchipNameBuf)
		nameParts := strings.Split(chipName, "-")

		adapter := C.GoString(C.sensors_get_adapter_name(&cchip.bus))

		chip := &Chip{ID: chipName, Adapter: adapter, Type: nameParts[0], Bus: nameParts[1], Address: nameParts[2], Sensors: map[string]Sensor{}}

		i := C.int(0)
		for {
			feature := C.sensors_get_features(cchip, &i)
			if feature == nil {
				break
			}
			sensorType := LmSensorType(feature._type)

			clabel := C.sensors_get_label(cchip, feature)
			if clabel == nil {
				continue
			}
			label := C.GoString(clabel)

			var reading Sensor

			switch sensorType {
			case Temperature:
				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_TEMP_INPUT)
				if sf != nil {
					value, err := getValue(cchip, sf)
					if err == nil {
						reading = &TempSensor{baseSensor{label, value}, Unknown}
					}
				}

				sf = C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_TEMP_TYPE)
				if reading != nil && sf != nil {
					value, err := getValue(cchip, sf)
					if err == nil {
						(reading.(*TempSensor)).TempType = LmTempType(int(value))
					}
				}

			case Voltage:
				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_IN_INPUT)
				if sf != nil {
					value, err := getValue(cchip, sf)
					if err == nil {
						reading = &VoltageSensor{baseSensor{label, value}}
					}
				}

			case Fan:
				sf := C.sensors_get_subfeature(cchip, feature, C.SENSORS_SUBFEATURE_FAN_INPUT)
				if sf != nil {
					value, err := getValue(cchip, sf)
					if err == nil {
						reading = &FanSensor{baseSensor{label, value}}
					}
				}

			default:
				reading = &UnimplementedSensor{baseSensor{Name: label}, sensorType}
			}

			if reading != nil {
				chip.Sensors[reading.GetName()] = reading
			}
		}

		sensors.Chips[chip.ID] = chip
	}

	return sensors, nil
}
