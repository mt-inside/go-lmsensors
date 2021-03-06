# go-lmsensors
Linux hardware sensor monitoring in Go.

[![Checks](https://github.com/mt-inside/go-lmsensors/actions/workflows/checks.yaml/badge.svg)](https://github.com/mt-inside/go-lmsensors/actions/workflows/checks.yaml)
[![GitHub Issues](https://img.shields.io/github/issues-raw/mt-inside/go-lmsensors)](https://github.com/mt-inside/go-lmsensors/issues)

[![Go Reference](https://pkg.go.dev/badge/github.com/mt-inside/go-lmsensors.svg)](https://pkg.go.dev/github.com/mt-inside/go-lmsensors)

Uses the [lm-sensors](https://github.com/lm-sensors/lm-sensors) (linux monitoring sensors) pacakge, on top of the [hwmon](https://hwmon.wiki.kernel.org) kernel feature.

## Setup
* Install _lm-sensors_
  * Ubuntu: `sudo apt install lm-sensors libsensors-dev`
  * Arch: `pacman -S lm_sensors`
* Configure _lm-sensors_
  * Run `sensors-detect`
  * Made any [necessary adjustments](https://hwmon.wiki.kernel.org/faq) to the [configuration](https://linux.die.net/man/5/sensors3.conf) in `/etc/sensors3.conf`, using `/etc/sensors.d/*`
* `go get github.com/mt-inside/go-lmsensors`

## How it works
This package links against the C-language `libsensors` and calls it to get sensor readings from the hwmon kernel subsystem (which it reads from sysfs).

My original version ran and parsed `sensors -j`, and all the information is in that JSON if you really squint and know how to read it.
However, using the library direct seemed faster, avoids a fork(), and doesn't require `lm-sensors` to be installed, just `libsensors5` (some package managers have them separately). (The instructions say to install lm-sensors, becuase you almost certainly want to run `sensors-detect`.)

The hwmon data _are_ exposed through sysfs, but those are raw values - libsensors isn't just a convenience binding; it scales raw values according to a big built-in database, and lets the user rename sensors.

## Example

### Code
```go
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

	for _, chip := range sensors.Chips {
		fmt.Println(chip.ID)
		for _, reading := range chip.Readings {
			fmt.Printf("  [%s] %s: %f\n", reading.SensorType, reading.Name, reading.Value)
		}
	}
}
```

### Output
```
it8792-isa-0a60
  [In] PM_CLDO12: 1.504000
  [Fan] SYS_FAN4: 0.000000
  [In] VIN0: 1.788000
  [In] DDR VTT: 0.665000
  [In] Chipset Core: 1.090000
  [In] six: 2.780000
  [Temp] PCIEX4_1: 37.000000
  [Temp] System2: 34.000000
...
```
