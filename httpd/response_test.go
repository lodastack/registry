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
		t.Fatalf("response of WriteJson not with expect, w: %+v, body struct: %+v, err:%v", *w, response, err)
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

	req := httptest.NewRequest("GET", "http://loda.com/test", nil)
	w := httptest.NewRecorder()
	handlerOK(w, req)
	if w.Code != 200 {
		t.Fatal("WriteOK return http code not 200")
	}

	w = httptest.NewRecorder()
	handlerBadRequest(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatal("WriteBadRequest return http code not 400")
	}

	w = httptest.NewRecorder()
	handlerServerError(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatal("WriteServerError return http code not 500")
	}
}
