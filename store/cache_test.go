package store

import (
	"testing"
)

var bucket = []byte("test-bucket")
var key = []byte("test-key")
var value = []byte("test-value123")

var bucket2 = []byte("test-bucket2")
var key2 = []byte("test-key2")
var value2 = []byte("test-value456")

var bucket3 = []byte("test-bucket3")
var key3 = []byte("test-key3")
var value3 = []byte("test-value789")

func TestCache(t *testing.T) {
	c, err := New(10, nil)
	if err != nil {
		t.Fatalf("new cache err: %v", err)
	}

	if c.maxSize != 10 {
		t.Fatalf("new cache err, expect maxSize: %d - %d", 10, c.maxSize)
	}
}

func Test_Add_Get(t *testing.T) {
	c, err := New(1024, nil)
	if err != nil {
		t.Fatalf("new cache err: %v", err)
	}

	c.Add(bucket, key, value)
	c.Add(bucket2, key2, value2)

	v, exist := c.Get(bucket, key)
	if !exist || string(v) != string(value) {
		t.Fatalf("Add Key err: %v %s - %s ", err, string(v), string(value))
	}

	v, exist = c.Get(bucket2, key2)
	if !exist || string(v) != string(value2) {
		t.Fatalf("Add Key err: %v %s - %s ", err, string(v), string(value2))
	}
}

func Test_LRU_FullMem(t *testing.T) {
	c, err := New(40, nil)
	if err != nil {
		t.Fatalf("new cache err: %v", err)
	}

	c.Add(bucket, key, value)
	c.Add(bucket2, key2, value2)
	c.Add(bucket3, key3, value3)

	v, exist := c.Get(bucket, key)
	if exist || string(v) == string(value) {
		t.Fatalf("key should be evicted")
	}

	v, exist = c.Get(bucket2, key2)
	if exist || string(v) == string(value2) {
		t.Fatalf("key should be evicted")
	}

	v, exist = c.Get(bucket3, key3)
	if !exist || string(v) != string(value3) {
		t.Fatalf("Get Key err: %v %s - %s ", err, string(v), string(value3))
	}
}

func Test_LRU(t *testing.T) {
	c, err := New(70, nil)
	if err != nil {
		t.Fatalf("new cache err: %v", err)
	}

	c.Add(bucket, key, value)
	c.Add(bucket2, key2, value2)
	c.Add(bucket3, key3, value3)

	c.Get(bucket2, key2)

	c.Add(bucket, key2, value2)

	c.Get(bucket2, key2)

	c.Add(bucket2, key, value)

	v, exist := c.Get(bucket, key)
	if exist || string(v) == string(value) {
		t.Fatalf("key should be evicted")
	}

	v, exist = c.Get(bucket3, key3)
	if exist || string(v) == string(value3) {
		t.Fatalf("key should be evicted")
	}

	v, exist = c.Get(bucket2, key2)
	if !exist || string(v) != string(value2) {
		t.Fatalf("Get Key err: %v %s - %s ", err, string(v), string(value2))
	}
}

func Test_RemoveBucket(t *testing.T) {
	c, err := New(1024, nil)
	if err != nil {
		t.Fatalf("new cache err: %v", err)
	}

	c.Add(bucket, key, value)
	c.Add(bucket2, key2, value2)
	c.Add(bucket3, key3, value3)

	c.RemoveBucket(bucket)

	v, exist := c.Get(bucket, key)
	if exist || string(v) == string(value) {
		t.Fatalf("key should be removed")
	}

	v, exist = c.Get(bucket2, key2)
	if !exist || string(v) != string(value2) {
		t.Fatalf("Get Key err: %v %s - %s ", err, string(v), string(value2))
	}

	v, exist = c.Get(bucket3, key3)
	if !exist || string(v) != string(value3) {
		t.Fatalf("Get Key err: %v %s - %s ", err, string(v), string(value3))
	}
}

func Test_Remove(t *testing.T) {
	c, err := New(1024, nil)
	if err != nil {
		t.Fatalf("new cache err: %v", err)
	}

	c.Add(bucket, key, value)
	c.Add(bucket, key2, value2)
	c.Add(bucket2, key2, value2)
	c.Add(bucket3, key3, value3)

	c.Remove(bucket, key)

	v, exist := c.Get(bucket, key)
	if exist || string(v) == string(value) {
		t.Fatalf("key should be removed")
	}

	v, exist = c.Get(bucket, key2)
	if !exist || string(v) != string(value2) {
		t.Fatalf("Get Key err: %v %s - %s ", err, string(v), string(value2))
	}

	v, exist = c.Get(bucket2, key2)
	if !exist || string(v) != string(value2) {
		t.Fatalf("Get Key err: %v %s - %s ", err, string(v), string(value2))
	}

	v, exist = c.Get(bucket3, key3)
	if !exist || string(v) != string(value3) {
		t.Fatalf("Get Key err: %v %s - %s ", err, string(v), string(value3))
	}
}

func Test_Purge(t *testing.T) {
	c, err := New(40, nil)
	if err != nil {
		t.Fatalf("new cache err: %v", err)
	}

	c.Add(bucket, key, value)
	c.Add(bucket2, key2, value2)
	c.Add(bucket3, key3, value3)

	c.Purge()

	v, exist := c.Get(bucket, key)
	if exist || string(v) == string(value) {
		t.Fatalf("key should be evicted")
	}

	v, exist = c.Get(bucket2, key2)
	if exist || string(v) == string(value2) {
		t.Fatalf("key should be evicted")
	}

	v, exist = c.Get(bucket3, key3)
	if exist || string(v) == string(value3) {
		t.Fatalf("key should be evicted")
	}
}
