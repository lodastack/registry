package common

import (
	"encoding/json"

	"github.com/lodastack/models"
	"github.com/lodastack/sdk-go"
)

func Send(ns string, ms []models.Metric) error {
	data, err := json.Marshal(ms)
	if err != nil {
		return err
	}
	return sdk.Post(ns, data)
}
