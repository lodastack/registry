package model

import (
	"testing"
)

var boltByte = []byte{72, 0, 72, 72, 101, 108, 108, 111, 111, 1, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 1, 1, 72, 72, 101, 108, 108, 111, 1, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 1, 1, 1,
	73, 0, 72, 101, 108, 108, 111, 111, 1, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 1, 1, 72, 101, 108, 108, 111, 1, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 2}

var rByte = []byte{72, 73, 74, 0, 72, 72, 101, 108, 108, 111, 111, 1, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 1, 1,
	72, 72, 101, 108, 108, 111, 1, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 2}

var resMaps []map[string]string = []map[string]string{{"res_key1": "res1_v1", "res_key2": "res1_v2"}, {"res_key1": "res2_v1", "res_key2": "res2_v2", "_id": "uuid1"}}
var resByte = []byte{91, 123, 34, 114, 101, 115, 95, 107, 101, 121, 49, 34, 58, 32, 34, 114, 101, 115, 49, 95, 118, 49, 34, 44, 32, 34, 114,
	101, 115, 95, 107, 101, 121, 50, 34, 58, 32, 34, 114, 101, 115, 49, 95, 118, 50, 34, 125, 44, 32, 123, 34, 114, 101, 115, 95,
	107, 101, 121, 49, 34, 58, 32, 34, 114, 101, 115, 50, 95, 118, 49, 34, 44, 32, 34, 114, 101, 115, 95, 107, 101, 121, 50, 34,
	58, 32, 34, 114, 101, 115, 50, 95, 118, 50, 34, 44, 32, 34, 95, 105, 100, 34, 58, 32, 34, 117, 117, 105, 100, 49, 34, 125, 93}

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
		t.Log("%+v", resouce)
		if _, ok := resouce["_id"]; !ok || len(resouce) != 3 {
			t.Fatalf("unmarshal fail, resource should have _id")
		}
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
	for _, resouce := range *ressStruct {
		_, havKey1 := resouce["res_key1"]
		_, havKey2 := resouce["res_key2"]
		if !havKey1 || !havKey2 {
			t.Fatalf("unmarshal fail, resource should have _id")
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

func TestLoadByte(t *testing.T) {
	ressStruct, err := NewResources(resByte)
	if err != nil {
		t.Fatalf("Resources load byte fail", err.Error())
	}
	if len(*ressStruct) != 2 {
		t.Fatalf("Resources load byte error: num of resource(map) not match")
	}
	for _, res := range *ressStruct {
		_, key1 := res["res_key1"]
		_, key2 := res["res_key2"]
		if !key1 || !key2 {
			t.Fatalf("Resources load byte error: property of resource(map) not match")
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
