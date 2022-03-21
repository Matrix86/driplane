package utils

import (
	"testing"
	"time"
)

func TestGetGlobalTTLMapInstance(t *testing.T) {
	m1 := GetGlobalTTLMapInstance(1 * time.Second)
	m2 := GetGlobalTTLMapInstance(1 * time.Second)
	if m1 != m2 {
		t.Errorf("the instances returned are different")
	}
}

func TestGetNamedTTLMap(t *testing.T) {
	m1 := GetNamedTTLMap("test", 1*time.Second)
	m2 := GetNamedTTLMap("test", 1*time.Second)
	m3 := GetNamedTTLMap("test2", 1*time.Second)

	if m1 != m2 {
		t.Errorf("the instances returned are different")
	}
	if m1 == m3 {
		t.Errorf("the instances returned should be different")
	}
}
