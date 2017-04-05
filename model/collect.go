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

func GenCollectName(res Resource) string {
	if res["measurement_type"] == PortCollect {
		return res["measurement_type"] + "." + res["name"] + "." + res["port"]
	}

	return res["measurement_type"] + "." + res["name"]
}

// collectTypeIllegle return true if collect type is illegal.
func collectTypeIllegal(res Resource) bool {
	collectType, _ := res["measurement_type"]
	if collectType == "" {
		return true
	}
	switch collectType {
	case PortCollect:
		if port, _ := res["port"]; port == "" {
			return true
		}
	}
	return false
}

func UpdateCollectName(collects ...Resource) error {
	for index := range collects {
		if collectTypeIllegal(collects[index]) {
			return ErrInvalidParam
		}

		// do not update name if the collect resource is base system collect.
		collectType := collects[index]["measurement_type"]
		if collectType != PortCollect &&
			collectType != ProcCollect &&
			collectType != PluginCollect &&
			collectType != ApiCollect {
			continue
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

		collects[index]["name"] = GenCollectName(collects[index])
	}
	return nil
}
