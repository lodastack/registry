package model

import (
	"github.com/lodastack/models"
)

type Report models.Report

var (
	HostnameProp   = "hostname"
	HostStatusProp = "status"
	IpProp         = "ip"

	Online  = "online"
	Offline = "offline"
	Dead    = "dead"
)
