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
	m.Put("key", "value", 1)
	l := len(m.dict)
	time.Sleep(2*time.Second)
	if _, ok := m.dict["key"]; ok {
		t.Errorf("the expired key has not been removed")
	}
	if len(m.dict) >= l {
		t.Errorf("the expired key has not been removed from the GC")
	}
}

func TestDeleteOnGet(t *testing.T) {
	ttl := 1*time.Second
	gc := 10*time.Second
	m := NewTTLMap(gc)
	m.Put("key", "value", int64(ttl.Seconds()))
	time.Sleep(2*time.Second)
	if _, ok := m.dict["key"]; !ok {
		t.Errorf("key should be still here")
	}
	if _, ok := m.Get("key"); ok {
		t.Errorf("the expired key has not been removed")
	}
}