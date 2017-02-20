package model

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/lodastack/registry/common"

	"github.com/lodastack/models"
)

var (
	VersionSep = "__"
	DbPrefix   = "collect."
)

type AlarmResource models.Alarm

func NewAlarm(ns, name string) *AlarmResource {
	return &AlarmResource{
		Name:    name,
		DB:      DbPrefix + ns,
		Enable:  "true",
		Default: "false"}
}

func (a *AlarmResource) SetID(id string) {
	a.ID = id
}

func (a *AlarmResource) SetQuery(function, rp, measurement, period, where,
	expression, every, groupby, trigger, shift, value string) error {
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
	return nil
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

func (a *AlarmResource) SetDefault() {
	a.Default = "true"
}

func (a *AlarmResource) UnsetDefault() {
	a.Default = "false"
}

func (a *AlarmResource) SetMD5AndVersion() error {
	if a.ID == "" {
		return errors.New("invalid id")
	}

	if len(a.DB) < len(DbPrefix)+1 {
		return errors.New("invalid db")
	}
	a.MD5, a.Version = "", ""
	ns := a.DB[len(DbPrefix):]

	md5, err := a.calMD5()
	if err != nil {
		return err
	}
	a.MD5 = md5
	a.Version = ns + VersionSep + a.Measurement + VersionSep + a.ID + VersionSep + md5
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

	function, OKFunction := data["function"]
	rp, OKRp := data["rp"]
	measurement, OKMeasurement := data["measurement"]
	period, OKPeriod := data["period"]
	where, _ := data["where"]
	expression, OKExpression := data["expression"]
	every, OKEvery := data["every"]
	groupby, OKGroupby := data["groupby"]
	trigger, OKTrigger := data["trigger"]
	shift, OKTrigger := data["shift"]
	value, OKValue := data["value"]

	level, OKLevel := data["level"]
	groups, OKGroups := data["groups"]
	alert, OKAlert := data["alert"]
	message, OKMessage := data["message"]

	if !OKFunction || !OKRp || !OKMeasurement ||
		!OKPeriod || !OKPeriod ||
		!OKExpression || !OKEvery || !OKGroupby ||
		!OKTrigger || !OKTrigger || !OKValue ||
		!OKLevel || !OKGroups || !OKAlert || !OKMessage {
		return &AlarmResource{}, ErrInvalidParam
	}

	if err := alarm.SetQuery(function, rp, measurement, period,
		where, expression, every, groupby, trigger, shift, value); err != nil {
		return alarm, err
	}
	if err := alarm.SetAlert(level, groups, alert, message); err != nil {
		return alarm, err
	}
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
