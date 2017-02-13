package utils

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	Form = "form"
	Raw  = "raw"
)

type HttpQuery struct {
	Timeout  int
	Method   string
	Url      string
	BodyType string
	Body     []byte
	Result   HttpResult
}

type HttpResult struct {
	Status int
	Body   []byte
}

func (query *HttpQuery) DoQuery() error {
	url, err := url.Parse(query.Url)
	if err != nil {
		return err
	}

	TimeoutDuration := time.Duration(query.Timeout) * time.Second
	client := &http.Client{Timeout: time.Duration(TimeoutDuration)}
	req, err := http.NewRequest(query.Method, url.String(), bytes.NewBufferString(string(query.Body)))
	if err != nil {
		return err
	}

	if query.Method != http.MethodGet {
		queryHeader := ""
		if query.BodyType == Form {
			queryHeader = "application/x-www-form-urlencoded"
		} else if query.BodyType != Raw {
			return errors.New("Unkown request body type")
		}
		req.Header.Set("Content-Type", queryHeader)
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	query.Result.Status = res.StatusCode
	query.Result.Body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New("error in read post body")
	}
	return nil
}
