package model

import (
	"testing"
)

// [map[_id:uuid1 res_key1:res1_v1 res_key2:res1_v2] map[_id:uuid2 res_key1:res2_v1 res_key2:res2_v2]]
var searchByte = []byte{117, 117, 105, 100, 49, 0, 114, 101, 115, 95, 107, 101, 121, 49, 1, 114, 101, 115, 49, 95, 118, 49, 1, 1,
	114, 101, 115, 95, 107, 101, 121, 50, 1, 114, 101, 115, 49, 95, 118, 50, 1, 1, 1,
	117, 117, 105, 100, 50, 0, 114, 101, 115, 95, 107, 101, 121, 49, 1, 114, 101, 115, 50, 95, 118, 49, 1, 1,
	114, 101, 115, 95, 107, 101, 121, 50, 1, 114, 101, 115, 50, 95, 118, 50, 2}

func TestIdSearch(t *testing.T) {
	search := ResourceSearch{
		Id: "uuid2",
	}
	search.Init()
	ressStruct := Resources{}
	if err := ressStruct.Unmarshal(searchByte); err != nil {
		t.Fatalf("Resources load byte fail: %s", err.Error())
	}
	t.Log(ressStruct)
	result, err := search.Process(searchByte)
	t.Log("search id uuid2 result:", result)
	if err != nil || len(result) == 0 || result[0]["_id"] != "uuid2" {
		t.Fatal("id search result not match: ", err)
	}
}

func TestValueSearchEmptyKey(t *testing.T) {
	search := ResourceSearch{
		Value: []byte("res1_v2"),
	}
	search.Init()
	ressStruct := Resources{}
	if err := ressStruct.Unmarshal(searchByte); err != nil {
		t.Fatalf("Resources load byte fail: %s", err.Error())
	}
	t.Log(ressStruct)
	result, err := search.Process(searchByte)
	t.Log("search v uuid2 result:", result)
	if err != nil || len(result) == 0 || result[0]["_id"] != "uuid1" || result[0]["res_k2"] == "res1_v2" {
		t.Fatal("value search result not match: ", err)
	}
}

func TestValueSearchHasKey(t *testing.T) {
	ressStruct := Resources{}
	if err := ressStruct.Unmarshal(searchByte); err != nil {
		t.Fatalf("Resources load byte fail: %s", err.Error())
	}
	t.Log(ressStruct)

	// case 1
	search := ResourceSearch{
		Key:   "res_key2",
		Value: []byte("res2_v2"),
	}
	search.Init()
	result, err := search.Process(searchByte)
	t.Log("search k-v uuid2 result:", result)
	if err != nil || len(result) == 0 || result[0]["_id"] != "uuid2" || result[0]["res_k2"] == "res2_v2" {
		t.Fatal("key-value search result not match: ", err)
	}

	// case 2
	search = ResourceSearch{
		Key:   "res_key2",
		Value: []byte("res1_v2"),
	}
	search.Init()
	result, err = search.Process(searchByte)
	t.Log("search k-v uuid2 result:", result)
	if err != nil || len(result) == 0 {
		t.Fatal("key-value search result not match: ", err)
	}
	if result[0]["_id"] != "uuid1" || result[0]["res_k2"] == "res1_v2" {
		t.Fatal("key-value search result not match: ", err)
	}

	// case 3
	search = ResourceSearch{
		Key:   "res_key3",
		Value: []byte("res2_v2"),
	}
	search.Init()
	result, err = search.Process(searchByte)
	t.Log("search k-v uuid2 result:", result)
	if len(result) != 0 {
		t.Fatal("key-value search result not match: ", err)
	}

	// case 4
	search = ResourceSearch{
		Key:   "res_key2",
		Value: []byte("res2_v3"),
	}
	search.Init()
	result, err = search.Process(searchByte)
	t.Log("search k-v uuid2 result:", result)
	if len(result) != 0 {
		t.Fatal("key-value search result not match: ", err)
	}
}
