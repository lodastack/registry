package model

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/lodastack/registry/common"

	"github.com/lodastack/models"
)

type AlarmResource models.Alarm

var (
	rp string = "loda"
)

func NewAlarm(ns, name string) *AlarmResource {
	return &AlarmResource{
		Name:    name,
		DB:      models.DBPrefix + ns,
		Enable:  "true",
		Default: "false"}
}

func (a *AlarmResource) SetID(id string) {
	a.ID = id
}

func (a *AlarmResource) SetQuery(function, rp, measurement, period, where,
	expression, every, groupby, trigger, shift, value, stime, etime string) error {
	a.Func = function
	a.RP = rp
	a.Measurement = measurement
	a.Where = where
	a.Expression = expression
	a.Period = period
	a.Every = every
	a.Trigger = trigger
	a.Shift = shift
	a.Value = value
	a.GroupBy = groupby
	a.STime = stime
	a.ETime = etime
	return nil
}

func (a *AlarmResource) SetBlock(data Resource) {
	if blockStep, ok := data["blockstep"]; !ok || blockStep == "" {
		a.BlockStep = "5"
	} else {
		a.BlockStep = blockStep
	}

	if maxBlockTime, ok := data["maxblocktime"]; !ok || maxBlockTime == "" {
		a.MaxBlockTime = "60"
	} else {
		a.MaxBlockTime = maxBlockTime
	}
}

func (a *AlarmResource) SetAlert(level, groups, alert, message string) error {
	a.Level = level
	a.Groups = groups
	a.Alert = alert
	a.Message = message
	return nil
}

func (a *AlarmResource) EnableSelf() {
	a.Enable = "true"
}

func (a *AlarmResource) DisableSelf() {
	a.Enable = "false"
}

func (a *AlarmResource) SetMD5AndVersion() error {
	if a.ID == "" {
		return errors.New("invalid id")
	}

	if len(a.DB) < len(models.DBPrefix)+1 {
		return errors.New("invalid db")
	}
	a.MD5, a.Version = "", ""
	ns := a.DB[len(models.DBPrefix):]

	md5, err := a.calMD5()
	if err != nil {
		return err
	}
	a.MD5 = md5
	a.Version = ns + models.VersionSep + a.Measurement + models.VersionSep + a.ID + models.VersionSep + md5
	return nil
}

func (a *AlarmResource) calMD5() (string, error) {
	md5Ctx := md5.New()
	bytes, err := json.Marshal(*a)
	if err != nil {
		return "", err
	}
	md5Ctx.Write(bytes)
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr), nil
}

func NewAlarmByRes(ns string, data Resource, ID string) (*AlarmResource, error) {
	name, ok := data["name"]
	if !ok || ns == "" {
		return &AlarmResource{}, ErrInvalidParam
	}

	alarm := NewAlarm(ns, name)
	if ID != "" {
		alarm.SetID(ID)
	} else {
		alarm.SetID(common.GenUUID())
	}
	if enable, _ := data["enable"]; enable == "false" {
		alarm.DisableSelf()
	}

	function, _ := data["func"]
	measurement, _ := data["measurement"]
	period, _ := data["period"]
	where, _ := data["where"]
	expression, _ := data["expression"]
	every, _ := data["every"]
	groupby, _ := data["groupby"]
	trigger, _ := data["trigger"]
	shift, _ := data["shift"]
	value, _ := data["value"]

	level, _ := data["level"]
	groups, _ := data["groups"]
	alert, _ := data["alert"]
	message, _ := data["message"]

	stime, _ := data["starttime"]
	etime, _ := data["endtime"]

	if measurement == "" || period == "" ||
		every == "" || trigger == "" || level == "" ||
		alert == "" || function == "" || groups == "" {
		return &AlarmResource{}, ErrInvalidParam
	}

	if (trigger == models.ThresHold && (value == "" || expression == "")) ||
		(trigger == models.Relative && (value == "" || expression == "")) {
		return &AlarmResource{}, ErrInvalidParam
	}

	if err := alarm.SetQuery(function, rp, measurement, period,
		where, expression, every, groupby, trigger, shift, value, stime, etime); err != nil {
		return alarm, err
	}
	if err := alarm.SetAlert(level, groups, alert, message); err != nil {
		return alarm, err
	}
	alarm.SetBlock(data)
	if err := alarm.SetMD5AndVersion(); err != nil {
		return alarm, err
	}

	return alarm, nil
}

func TransAlarmToResource(alarm AlarmResource) (Resource, error) {
	mapByte, err := json.Marshal(alarm)
	if err != nil {
		return Resource{}, err
	}
	mapData := map[string]string{}
	err = json.Unmarshal(mapByte, &mapData)
	if err != nil {
		return Resource{}, err
	}
	return NewResource(mapData), nil
}

func NewAlarmResourceByMap(ns string, data map[string]string, ID string) (Resource, error) {
	alarm, err := NewAlarmByRes(ns, data, ID)
	if err != nil {
		return nil, err
	}

	return TransAlarmToResource(*alarm)
}
