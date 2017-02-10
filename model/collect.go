package model

import (
	"strings"
)

var (
	ProcCollect   = "PROC"
	PluginCollect = "PLUGIN"
	PortCollect   = "PORT"
)

// GetNameFromMeasurements get the resource names of the measurements.
// PROC.bin.cpu.idle -> PROC.bin
// PLUGIN.name.cpu.idle -> PLUGIN.name
// PORT.service.xx -> PORT.service.xx
func GetResNameFromMeasurements(measurements []string) ([]string, bool) {
	resNames := make([]string, len(measurements))
	cnt := 0
	for _, measurement := range measurements {
		nameSplit := strings.Split(measurement, ".")
		if len(nameSplit) < 2 {
			continue
		}
		switch nameSplit[0] {
		case ProcCollect:
			fallthrough
		case PluginCollect:
			if len(nameSplit) > 2 {
				resNames[cnt] = strings.Join(nameSplit[:2], ".")
			} else if len(nameSplit) == 2 {
				resNames[cnt] = measurement
			}
		case PortCollect:
			resNames[cnt] = measurement

		default:
			continue
		}
		cnt++
	}

	if cnt == 0 {
		return nil, false
	}
	return resNames[:cnt], true
}

func GenCollectName(collectType, collectName string) string {
	return collectType + "." + collectName
}

// collectTypeIllegle return true if collect type is illegal.
func collectTypeIllegal(collectType string) bool {
	if collectType == "" {
		return true
	}
	return false
}

func UpdateCollectName(collects ...Resource) bool {
	for index := range collects {
		collectType, _ := collects[index]["measurement_type"]
		if collectTypeIllegal(collectType) {
			return false
		}
		collects[index]["name"] = GenCollectName(collectType, collects[index]["name"])
	}
	return true
}
