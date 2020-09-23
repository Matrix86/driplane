package utils

import (
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
		value:      "value",
		expiration: time.Now().Unix(),
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
		value:      "value",
		expiration: time.Now().Unix() -1,
	}
	if _, ok := m.dict["key"]; !ok {
		t.Errorf("key should be still here")
	}
	if _, ok := m.Get("key"); ok {
		t.Errorf("the expired key has not been removed")
	}

	m.dict["key"] = &item{
		value:      "value",
		expiration: time.Now().Unix() +2,
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
	if v.value != "value" {
		t.Errorf("wrong value")
	}

	// Overwriting
	m.Put("key", "newvalue", 10)
	if !ok {
		t.Errorf("key should be still here")
	}
	if v.value != "newvalue" {
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