package model

type Row struct {
	Key    []byte `json:"key,omitempty"`
	Value  []byte `json:"value,omitempty"`
	Bucket []byte `json:"bucket,omitempty"`
}
