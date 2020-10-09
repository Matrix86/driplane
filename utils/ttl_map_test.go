package utils

import (
	"os"
	"testing"
	"time"
)

func TestNewTTLMap(t *testing.T) {
	gc := 1*time.Second
	m := NewTTLMap(gc)
	if m.gcdelay != gc {
		t.Errorf("'gc' parameter is wrong")
	}
}

func TestDeleteOnGC(t *testing.T) {
	gc := 1*time.Second
	m := NewTTLMap(gc)
	m.Put("key", "value", 0)
	l := len(m.dict)
	time.Sleep(1001*time.Millisecond)
	if _, ok := m.dict["key"]; ok {
		t.Errorf("the expired key has not been removed")
	}
	if len(m.dict) >= l {
		t.Errorf("the expired key has not been removed from the GC")
	}
}

func TestDeleteOnGet(t *testing.T) {
	gc := 10*time.Second
	m := NewTTLMap(gc)
	m.dict["key"] = &item{
		Value:      "value",
		Expiration: time.Now().Unix(),
	}
	// Key should not be removed because the GC is going to do that every 10s
	if _, ok := m.dict["key"]; !ok {
		t.Errorf("key should be still here")
	}

	if _, ok := m.Get("key"); ok {
		t.Errorf("the expired key has not been removed")
	}
}

func TestTTLMap_Get(t *testing.T) {
	gc := 10*time.Second
	m := NewTTLMap(gc)
	m.dict["key"] = &item{
		Value:      "value",
		Expiration: time.Now().Unix() -1,
	}
	if _, ok := m.dict["key"]; !ok {
		t.Errorf("key should be still here")
	}
	if _, ok := m.Get("key"); ok {
		t.Errorf("the expired key has not been removed")
	}

	m.dict["key"] = &item{
		Value:      "value",
		Expiration: time.Now().Unix() +2,
	}
	v, ok := m.Get("key")
	if !ok {
		t.Errorf("the key should be still there")
	}
	if v != "value" {
		t.Errorf("the key has a wrong value")
	}

	_, ok = m.Get("notaddedkey")
	if ok {
		t.Errorf("the key should not be found")
	}
}

func TestTTLMap_Put(t *testing.T) {
	gc := 10*time.Second
	m := NewTTLMap(gc)
	m.Put("key", "value", 10)
	v, ok := m.dict["key"]
	if !ok {
		t.Errorf("key should be still here")
	}
	if v.Value != "value" {
		t.Errorf("wrong value")
	}

	// Overwriting
	m.Put("key", "newvalue", 10)
	if !ok {
		t.Errorf("key should be still here")
	}
	if v.Value != "newvalue" {
		t.Errorf("value has not been overwritten")
	}
}

func TestTTLMap_Len(t *testing.T) {
	gc := 10*time.Second
	m := NewTTLMap(gc)

	if m.Len() != 0 {
		t.Errorf("cache should be empty")
	}

	m.Put("key", "newvalue", 10)
	if m.Len() != 1 {
		t.Errorf("cache should contain a single value")
	}
}

func TestTTLMap_SetPersistence(t *testing.T) {
	gc := 1*time.Second
	m := NewTTLMap(gc)
	err := m.SetPersistence("/aaa/ttl_map_cache.data")
	if err == nil || err.Error() != "open /aaa/ttl_map_cache.data: no such file or directory" {
		t.Errorf("wrong error: %s", err)
	}
	filename := "/tmp/ttl_map_cache.data"
	err = m.SetPersistence(filename)
	if err != nil {
		t.Errorf("wrong error: %s", err)
	}
	err = m.SetPersistence(filename)
	if err != nil && err.Error() != "file already loaded: '/tmp/ttl_map_cache.data'" {
		t.Errorf("wrong error: expected=%s had=%s", "file already loaded: '/tmp/ttl_map_cache.data'", err)
	}
	if !FileExists(filename) {
		t.Errorf("file not created '/tmp/ttl_map_cache.data'")
	}
	defer os.Remove(filename)
	m.Put("test", "test", 10)
	err = m.syncFile()
	if err != nil {
		t.Errorf("wrong error on syncFile: expected=nil had=%s", err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Errorf("wrong error on os.Stat: expected=nil had=%s", err)
	}
	if info.Size() == 0 {
		t.Errorf("wrong file size: expected=%d had=%d", 1, info.Size())
	}

	m2 := NewTTLMap(gc)
	err = m2.SetPersistence(filename)
	if err != nil {
		t.Errorf("wrong error: %s", err)
	}
	v, ok := m2.Get("test")
	if ok && v.(string) != "test" {
		t.Errorf("wrong value: %s", v)
	}
}