package model

import (
	"github.com/lodastack/models"
)

type Report models.Report

var (
	HostnameProp   = "hostname"
	HostStatusProp = "status"
	IpProp         = "ip"
	SNProp         = "sn"
	SleepProp      = "sleep"

	Online  = "online"
	Offline = "offline"
	Dead    = "dead"
)
