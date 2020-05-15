package utils

import (
	"testing"
	"time"
)

func TestNewTTLMap(t *testing.T) {
	ttl := 1*time.Second
	gc := 1*time.Second
	m := NewTTLMap(ttl.Seconds(), gc)
	if m.ttl != int64(ttl.Seconds()) {
		t.Errorf("'ttl' parameter is wrong")
	}
	if m.gcdelay != gc {
		t.Errorf("'gc' parameter is wrong")
	}
}

func TestDeleteOnGC(t *testing.T) {
	ttl := 1*time.Second
	gc := 1*time.Second
	m := NewTTLMap(ttl.Seconds(), gc)
	m.Put("key", "value")
	time.Sleep(2*time.Second)
	if _, ok := m.dict["key"]; ok {
		t.Errorf("the expired key has not been removed")
	}
}

func TestDeleteOnGet(t *testing.T) {
	ttl := 1*time.Second
	gc := 10*time.Second
	m := NewTTLMap(ttl.Seconds(), gc)
	m.Put("key", "value")
	time.Sleep(2*time.Second)
	if _, ok := m.dict["key"]; !ok {
		t.Errorf("key should be still here")
	}
	if _, ok := m.Get("key"); ok {
		t.Errorf("the expired key has not been removed")
	}
}