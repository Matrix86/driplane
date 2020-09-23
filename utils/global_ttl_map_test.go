package utils

import (
	"testing"
	"time"
)

func TestGetGlobalTTLMapInstance(t *testing.T) {
	m1 := GetGlobalTTLMapInstance(1*time.Second)
	m2 := GetGlobalTTLMapInstance(1*time.Second)
	if m1 != m2 {
		t.Errorf("the instances returned are different")
	}
}
