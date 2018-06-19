package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lodastack/log"
	"github.com/lodastack/models"
	"github.com/lodastack/registry/common"
	"github.com/lodastack/registry/model/gitlab"
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
// RUN.API.Ping.xx -> Run.API.Ping.xx
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
		case RunPrefix:
			fallthrough
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

		// proc monitor: /sbin/python2.7, the collect name will be pythone2.7
		// we replace "." with "-".
		if collectType == ProcCollect {
			collects[index]["name"] = strings.Replace(collects[index]["name"], ".", "-", -1)
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

func getAlarmOfCollect(ns, collectType, collectName, groups string) ([]AlarmResource, error) {
	alarms := []AlarmResource{}
	var err error
	switch collectType {
	case PluginCollect:
		projectName := strings.Split(collectName, ".")[1]
		content, err := gitlab.GetFileContent(projectName)
		if err != nil {
			log.Warningf("read plugin %s alarm file error: %s", projectName, err.Error())
			return nil, nil
		}
		if err = json.Unmarshal([]byte(content), &alarms); err != nil {
			return nil, err
		}

	case PortCollect:
		fallthrough

	case ProcCollect:
		alarm := AlarmResource{}
		var measurement string
		if collectType == PortCollect {
			measurement = collectName
		} else {
			measurement = collectName + ".procnum"
		}
		alarm.Name = measurement + " change alert"
		alarm.Measurement = measurement
		alarm.Expression = ">"
		alarm.Trigger = models.Relative
		alarm.Value = "0"
		alarms = append(alarms, alarm)

	default:
		return nil, fmt.Errorf("unknow collect type to add alarm")
	}

	for i := range alarms {
		alarms[i].DB = models.DBPrefix + ns
		alarms[i].Enable, alarms[i].Default = "true", "false"
		alarms[i].BlockStep, alarms[i].MaxBlockTime = "10", "60"
		alarms[i].SetQuery("mean", rp, alarms[i].Measurement, "2m",
			"", alarms[i].Expression, "1m", "*", alarms[i].Trigger, "", alarms[i].Value, "0", "0")
		alarms[i].SetAlert("2", groups, "mail", "")
		alarms[i].SetID(common.GenUUID())
	}

	return alarms, err
}

func GetAlarmFromCollect(res Resource, ns, groups string) ([]Resource, error) {
	collectType, _ := res["measurement_type"]
	alarms, err := getAlarmOfCollect(ns, collectType, res["name"], groups)
	if err != nil || len(alarms) == 0 {
		return nil, err
	}

	for i := range alarms {
		if err := alarms[i].SetMD5AndVersion(); err != nil {
			return nil, err
		}
	}
	resourcelist := make([]Resource, len(alarms))
	for i, alarm := range alarms {
		resourcelist[i], err = TransAlarmToResource(alarm)
		if err != nil {
			return nil, err
		}
	}
	return resourcelist, nil
}
