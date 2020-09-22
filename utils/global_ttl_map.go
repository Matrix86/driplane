package utils

import (
	"sync"
	"time"
)

// GlobalTTLMap is a cache shared between all the rules
type GlobalTTLMap struct {
	Cache *TTLMap
}

var (
	instance *GlobalTTLMap
	once     sync.Once
)

// GetGlobalTTLMapInstance returns the unique GlobalTTLMap (singleton)
func GetGlobalTTLMapInstance(gcdelay time.Duration) *GlobalTTLMap {
	once.Do(func() {
		instance = &GlobalTTLMap{
			Cache: NewTTLMap(gcdelay),
		}
	})
	return instance
}
