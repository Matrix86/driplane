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
	mu       sync.Mutex
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

// GetNamedTTLMap return a Cache stored on the globalTTLMap with a name.
// A stored map that has already been closed is replaced with a fresh one:
// the "shutdown" event closes every cache, and that event is published on
// each reload too, so after a reload the registry would otherwise keep
// handing out an unusable map. The contents are not lost, because Close
// flushes them to the persistence file and the new map reloads it.
func GetNamedTTLMap(name string, gcdelay time.Duration) *TTLMap {
	i := GetGlobalTTLMapInstance(gcdelay)

	mu.Lock()
	defer mu.Unlock()

	if v, ok := i.Caches[name]; ok && !v.IsClosed() {
		return v
	}
	i.Caches[name] = NewTTLMap(gcdelay)
	return i.Caches[name]
}
