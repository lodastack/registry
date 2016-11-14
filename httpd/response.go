package httpd

import (
	"encoding/json"
	"errors"
	"net/http"
)

var errMarshalOutput = errors.New("Marshal JSON output fail.")

type Response struct {
	Code int
	Body []byte
	Json interface{}
}

func NewResponse(code int, body string, jsonStruct interface{}) Response {
	return Response{Code: code, Body: []byte(body), Json: jsonStruct}
}

func (r *Response) Write(w http.ResponseWriter) {
	if r.Code == 0 {
		r.Code = http.StatusOK
	}
	w.WriteHeader(r.Code)
	w.Write(r.Body)
}

// If marshal JSON fail, return 500.
func (r *Response) ReturnJson(w http.ResponseWriter) {
	var err error
	if r.Code == 0 {
		r.Code = http.StatusOK
	}
	r.Body, err = json.Marshal(r.Json)
	if err != nil {
		ReturnServerError(w, errMarshalOutput)
	} else {
		w.WriteHeader(r.Code)
		w.Write(r.Body)
	}
}

func WriteResponse(w http.ResponseWriter, code int, body []byte) {
	(&Response{Code: code, Body: body}).Write(w)
}

// Return 200 http status.
func ReturnOK(w http.ResponseWriter, body string) {
	WriteResponse(w, http.StatusOK, []byte(body))
}

// Return 400 http status.
func ReturnBadRequest(w http.ResponseWriter, err error) {
	WriteResponse(w, http.StatusBadRequest, []byte(err.Error()))
}

// Return 500 http status.
func ReturnServerError(w http.ResponseWriter, err error) {
	WriteResponse(w, http.StatusInternalServerError, []byte(err.Error()))
}

func ReturnJson(w http.ResponseWriter, httpStatus int, returnJson interface{}) {
	if httpStatus == 0 {
		httpStatus = http.StatusOK
	}
	(&Response{Code: httpStatus, Json: returnJson}).ReturnJson(w)
}
