package httpd

import (
	"encoding/json"
	"errors"
	"net/http"
)

var errMarshalOutput = errors.New("Marshal JSON output fail.")

type Response struct {
	Code int         `json:"httpstatus"`
	Data interface{} `json:"data"`
}

func (r *Response) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(r)
	if err != nil {
		ReturnServerError(w, errMarshalOutput)
	}
	if r.Code == 0 {
		r.Code = http.StatusOK
	}
	w.WriteHeader(r.Code)
	w.Write(body)
}

// Return 200 http status.
func ReturnOK(w http.ResponseWriter, msg string) {
	(&Response{Code: http.StatusOK, Data: msg}).Write(w)
}

// Return 404 http status.
func ReturnNotFound(w http.ResponseWriter, msg string) {
	(&Response{Code: http.StatusNotFound, Data: msg}).Write(w)
}

// Return 400 http status.
func ReturnBadRequest(w http.ResponseWriter, err error) {
	(&Response{Code: http.StatusBadRequest, Data: err.Error()}).Write(w)
}

// Return 500 http status.
func ReturnServerError(w http.ResponseWriter, err error) {
	(&Response{Code: http.StatusInternalServerError, Data: err.Error()}).Write(w)
}

func ReturnJson(w http.ResponseWriter, httpStatus int, returnJson interface{}) {
	if httpStatus == 0 {
		httpStatus = http.StatusOK
	}
	(&Response{Code: httpStatus, Data: returnJson}).Write(w)
}
