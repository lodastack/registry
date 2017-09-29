package test_sample

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func LoadFromFile(jsonFile string) ([]byte, error) {
	f, err := os.Open(jsonFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func LoadJsonFromFile(jsonFile string, v interface{}) error {
	bytes, err := LoadFromFile(jsonFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}
