package utils

import (
	"sync"
	"time"
)

// GlobalTTLMap is a cache shared between all the rules
type GlobalTTLMap struct {
	Caches map[string]*TTLMap
}

var (
	instance *GlobalTTLMap
	once     sync.Once
)

// GetGlobalTTLMapInstance returns the unique GlobalTTLMap (singleton)
func GetGlobalTTLMapInstance(gcdelay time.Duration) *GlobalTTLMap {
	once.Do(func() {
		instance = &GlobalTTLMap{
			Caches: make(map[string]*TTLMap),
		}
		instance.Caches["global"] = NewTTLMap(gcdelay)
	})
	return instance
}

// GetNamedTTLMap return a Cache stored on the globalTTLMap with a name
func GetNamedTTLMap(name string, gcdelay time.Duration) *TTLMap {
	i := GetGlobalTTLMapInstance(gcdelay)
	if v, ok := i.Caches[name]; ok {
		return v
	}
	i.Caches[name] = NewTTLMap(gcdelay)
	return i.Caches[name]
}
