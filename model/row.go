package model

import (
	"github.com/lodastack/log"
)

// log file backend
var LogBackend *log.FileBackend

type Row struct {
	Key    []byte `json:"key,omitempty"`
	Value  []byte `json:"value,omitempty"`
	Bucket []byte `json:"bucket,omitempty"`
}
