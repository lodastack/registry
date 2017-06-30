package gitlab

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lodastack/registry/config"
)

type gitUrl string

func (u *gitUrl) ToJSON(obj interface{}) error {
	body, err := get(string(*u))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, obj); err != nil {
		return err
	}
	return nil
}

func get(url string) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", config.C.PluginConf.Token)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
