package model

import (
	"fmt"
	"testing"
)

// [map[HHello: playground _id:H HHelloo: playground] map[_id:I Helloo: playground Hello: playgrou]]
var boltByte = []byte{72, 0, 72, 72, 101, 108, 108, 111, 111, 1, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 1, 1, 72, 72, 101, 108, 108, 111, 1, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 1, 1, 1,
	73, 0, 72, 101, 108, 108, 111, 111, 1, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 1, 1, 72, 101, 108, 108, 111, 1, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 2}

var rByte = []byte{72, 73, 74, 0, 72, 72, 101, 108, 108, 111, 111, 1, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 1, 1,
	72, 72, 101, 108, 108, 111, 1, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 2}

var resMaps []map[string]string = []map[string]string{{"res_key1": "res1_v1", "res_key2": "res1_v2"}, {"res_key1": "res2_v1", "res_key2": "res2_v2", "_id": "uuid1"}}
var resByte = []byte{91, 123, 34, 114, 101, 115, 95, 107, 101, 121, 49, 34, 58, 32, 34, 114, 101, 115, 49, 95, 118, 49, 34, 44, 32, 34, 114,
	101, 115, 95, 107, 101, 121, 50, 34, 58, 32, 34, 114, 101, 115, 49, 95, 118, 50, 34, 125, 44, 32, 123, 34, 114, 101, 115, 95,
	107, 101, 121, 49, 34, 58, 32, 34, 114, 101, 115, 50, 95, 118, 49, 34, 44, 32, 34, 114, 101, 115, 95, 107, 101, 121, 50, 34,
	58, 32, 34, 114, 101, 115, 50, 95, 118, 50, 34, 44, 32, 34, 95, 105, 100, 34, 58, 32, 34, 117, 117, 105, 100, 49, 34, 125, 93}

var emptyResRes []map[string]string = []map[string]string{{"res_key1": "", "res_key2": ""}, {"res_key1": "res2_v1", "res_key2": "", "_id": ""}}

func TestEmptyValueResource(t *testing.T) {
	res, err := NewResourcesMaps(emptyResRes)
	if err != nil {
		t.Fatalf("new resource from a map with empty value fail: %s", err.Error())
	}
	ressByte, err := res.Marshal()
	if err != nil {
		t.Fatalf("marshal a resource with empty property value fail: %s", err.Error())
	}
	*res = Resources{}
	err = res.Unmarshal(ressByte)
	if err != nil ||
		len(*res) != 2 ||
		len((*res)[0]) != 3 ||
		len((*res)[1]) != 3 ||
		(*res)[0]["res_key2"] != "" {
		t.Fatalf("unmarshal a resource byte with empty property value fail, len of resources: %d, len of rsource: %d %d, res_key2:%s, resources: %v, unmarshal error:%v", len(*res), len((*res)[0]), len((*res)[1]), (*res)[0]["res_key2"], *res, err)
	}

}

func TestRsUnmarshal(t *testing.T) {
	boltv := Resources{}
	if err := boltv.Unmarshal(boltByte); err != nil {
		t.Fatalf("unmarshal fail")
		return
	}
	t.Log(boltv, len(boltv))
	if len(boltv) != 2 {
		t.Fatalf("unmarshal fail, expect result of unmarshal have length: 2")
	}
	for _, resouce := range boltv {
		if _, ok := resouce["_id"]; !ok || len(resouce) != 3 {
			t.Fatalf("unmarshal fail, resource should have _id")
		}
		if resouce["_id"] == "H" {
			if v, ok := resouce["HHello"]; !ok || v != "playground" {
				t.Fatalf("unmarshal fail, resource not match with expect")
			}
			if v, ok := resouce["HHelloo"]; !ok || v != "playground" {
				t.Fatalf("unmarshal fail, resource not match with expect")
			}
		} else if resouce["_id"] == "I" {
			if v, ok := resouce["Hello"]; !ok || v != "playground" {
				t.Fatalf("unmarshal fail, resource not match with expect")
			}
			if v, ok := resouce["Helloo"]; !ok || v != "playground" {
				t.Fatalf("unmarshal fail, resource not match with expect, v is: %s", v)
			}
		}
	}
	if boltv[0]["_id"] == boltv[1]["_id"] {
		t.Fatalf("unmarshal fail, resource have same resource")
	}
}

func TestAppendResource(t *testing.T) {
	addRes := NewResource(map[string]string{"add_key1": "add_v1", "add_key2": "add_v2"})
	// addRes := Resource{}
	// addRes = addResMap

	newRsByte, _, err := AppendResources(boltByte, addRes)
	if err != nil {
		t.Fatalf("AppendResource fail: %s", err.Error())
	}

	newRs := Resources{}
	if err = newRs.Unmarshal(newRsByte); err != nil {
		t.Fatalf("unmarshal fail: %s", err.Error())
		return
	}
	if len(newRs) != 3 {
		t.Fatalf("unmarshal fail, expect result of unmarshal have length: 2")
	}
	for _, resouce := range newRs {
		if _, ok := resouce["_id"]; !ok || len(resouce) != 3 {
			t.Fatalf("unmarshal fail, resource should have _id")
		}
		if resouce["_id"] == "H" {
			if v, ok := resouce["HHello"]; !ok || v != "playground" {
				t.Fatalf("unmarshal fail, resource not match with expect")
			}
			if v, ok := resouce["HHelloo"]; !ok || v != "playground" {
				t.Fatalf("unmarshal fail, resource not match with expect")
			}
		} else if resouce["_id"] == "I" {
			if v, ok := resouce["Hello"]; !ok || v != "playground" {
				t.Fatalf("unmarshal fail, resource not match with expect")
			}
			if v, ok := resouce["Helloo"]; !ok || v != "playground" {
				t.Fatalf("unmarshal fail, resource not match with expect, v is: %s", v)
			}
		} else {
			if v, ok := resouce["add_key1"]; !ok || v != "add_v1" {
				t.Fatalf("unmarshal fail, resource not match with expect")
			}
			if v, ok := resouce["add_key2"]; !ok || v != "add_v2" {
				t.Fatalf("unmarshal fail, resource not match with expect, v is: %s", v)
			}
		}
	}
	if newRs[0]["_id"] == newRs[1]["_id"] {
		t.Fatalf("unmarshal fail, resource have same resource")
	}
}

func TestRsMarshal(t *testing.T) {
	ressStruct, err := NewResourcesMaps(resMaps)
	if err != nil {
		t.Fatalf("load map to  resources fail")
	}

	ressByte, err := ressStruct.Marshal()
	if err != nil {
		t.Fatalf("marshal resources fail")
	}
	*ressStruct = Resources{}
	if err := ressStruct.Unmarshal(ressByte); err != nil {
		t.Fatalf("unmarshal fail")
		return
	}
	t.Log(*ressStruct, len(*ressStruct))
	if len(*ressStruct) != 2 {
		t.Fatalf("unmarshal fail, expect result of unmarshal have length: 2")
	}
	for index, resouce := range *ressStruct {
		if v, k := resouce["res_key1"]; !k || v != fmt.Sprintf("res%d_v1", index+1) {
			t.Fatalf("unmarshal not match with expect, Unmarshal value is: %s", v)
		}
		if v, k := resouce["res_key2"]; !k || v != fmt.Sprintf("res%d_v2", index+1) {
			t.Fatalf("unmarshal not match with expect, Unmarshal value is: %s", v)
		}
	}
}

func TestRUnmarshal(t *testing.T) {
	r := Resource{}
	if err := r.Unmarshal(rByte); err != nil {
		t.Fatalf("unmarshal r fail: %s", err.Error())
	}
	if r[idKey] != "HIJ" || len(r) != 3 {
		t.Fatalf("unmarshal r fail: not match with expect")
	}
	t.Log(r)
}

func TestNewResources(t *testing.T) {
	ressStruct, err := NewResources(resByte)
	if err != nil {
		t.Fatalf("Resources load byte fail", err.Error())
	}
	if len(*ressStruct) != 2 {
		t.Fatalf("Resources load byte error: num of resource(map) not match")
	}
	for index, resouce := range *ressStruct {
		if v, k := resouce["res_key1"]; !k || v != fmt.Sprintf("res%d_v1", index+1) {
			t.Fatalf("unmarshal not match with expect, Unmarshal value is: %s", v)
		}
		if v, k := resouce["res_key2"]; !k || v != fmt.Sprintf("res%d_v2", index+1) {
			t.Fatalf("unmarshal not match with expect, Unmarshal value is: %s", v)
		}
	}

}

func BenchmarkUnmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		boltv := Resources{}
		boltv.Unmarshal(boltByte)
	}
}

func BenchmarkMarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ressStruct, _ := NewResourcesMaps(resMaps)
		ressStruct.Marshal()
	}
}
