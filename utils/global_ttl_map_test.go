package utils

import (
	"path/filepath"
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

// A named map that has been closed must not be handed out again: the shutdown
// event closes every cache and is published on each reload, so the registry
// would otherwise return an unusable map to the rebuilt filters.
func TestGetNamedTTLMapReplacesClosedMap(t *testing.T) {
	first := GetNamedTTLMap("reload-test", time.Second)
	if first.IsClosed() {
		t.Fatal("a fresh map should not be closed")
	}

	first.Close()
	if !first.IsClosed() {
		t.Fatal("Close() should mark the map as closed")
	}

	second := GetNamedTTLMap("reload-test", time.Second)
	if second == first {
		t.Fatal("the registry returned the closed map instead of a new one")
	}
	if second.IsClosed() {
		t.Fatal("the replacement map should be usable")
	}

	// and it must actually work, which is what the reload path needs
	if err := second.SetPersistence(filepath.Join(t.TempDir(), "c.dat")); err != nil {
		t.Fatalf("SetPersistence on the replacement map: %s", err)
	}
	second.Close()
}
