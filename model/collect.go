package model

import (
	"strings"

	"github.com/lodastack/log"
)

var (
	ProcCollect   = "PROC"
	PluginCollect = "PLUGIN"
	PortCollect   = "PORT"
	ApiCollect    = "API"

	RunPrefix = "RUN"
	RunType   = []string{ApiCollect}
)

// GetNameFromMeasurements get the resource names of the measurements.
// PROC.bin.cpu.idle -> PROC.bin
// PLUGIN.name.cpu.idle -> PLUGIN.name
// PORT.service.xx -> PORT.service.xx
// RUN.API.Ping.xx -> RUN.API.Ping
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
		case RunPrefix:
			if len(nameSplit) > 3 {
				resNames[cnt] = strings.Join(nameSplit[1:3], ".")
			} else {
				log.Errorf("invalid collect name %s, skip", measurement)
			}
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

func UpdateCollectName(collects ...Resource) error {
	for index := range collects {
		collectType, _ := collects[index]["measurement_type"]
		if collectTypeIllegal(collectType) {
			return ErrInvalidParam
		}

		for _, nameLetter := range collects[index]["name"] {
			if nameLetter == '-' ||
				(nameLetter >= 'a' && nameLetter <= 'z') ||
				(nameLetter >= 'A' && nameLetter <= 'Z') ||
				(nameLetter >= '0' && nameLetter <= '9') {
				continue
			}
			return ErrInvalidParam
		}
		collects[index]["name"] = GenCollectName(collectType, collects[index]["name"])
	}
	return nil
}
