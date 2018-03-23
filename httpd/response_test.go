package httpd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJson(t *testing.T) {
	type testResponse struct {
		Msg string
	}

	handlerJson := func(w http.ResponseWriter, r *http.Request) {
		result := testResponse{Msg: "test msg"}
		ReturnJson(w, http.StatusOK, result)
	}
	req := httptest.NewRequest("GET", "http://loda.com/test", nil)
	w := httptest.NewRecorder()
	handlerJson(w, req)

	response := Response{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if w.Code != 200 || err != nil || response.Data.(map[string]interface{})["Msg"].(string) != "test msg" {
		t.Fatalf("response of WriteJson not with expect, w: %+v, body struct: %+v, err:%v\n", *w, response, err)
	}
}

func TestWriteStatus(t *testing.T) {
	handlerOK := func(w http.ResponseWriter, r *http.Request) {
		ReturnOK(w, "test pass")
	}

	handlerBadRequest := func(w http.ResponseWriter, r *http.Request) {
		ReturnBadRequest(w, fmt.Errorf("test bad request"))
	}

	handlerServerError := func(w http.ResponseWriter, r *http.Request) {
		ReturnServerError(w, fmt.Errorf("test server error"))
	}

	var resp Response
	req := httptest.NewRequest("GET", "http://loda.com/test", nil)
	w := httptest.NewRecorder()
	handlerOK(w, req)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response fail: %s\n", err.Error())
	}
	if w.Code != 200 || resp.Msg != "test pass" {
		t.Fatalf("ReturnOK return not match with expect,code: %d, resp: %+v\n", w.Code, resp)
	}

	w = httptest.NewRecorder()
	handlerBadRequest(w, req)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response fail: %s\n", err.Error())
	}
	if w.Code != http.StatusBadRequest || resp.Msg != "test bad request" {
		t.Fatalf("ReturnBadRequest return not match with expect,code: %d, resp: %+v\n", w.Code, resp)
	}

	w = httptest.NewRecorder()
	handlerServerError(w, req)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response fail: %s\n", err.Error())
	}
	if w.Code != http.StatusInternalServerError || resp.Msg != "test server error" {
		t.Fatalf("ReturnServerError return not match with expect,code: %d, resp: %+v\n", w.Code, resp)
	}
}
