package utils

import (
	"sync"
	"time"
)

type GlobalTTLMap struct {
	Cache *TTLMap
}

var (
	instance *GlobalTTLMap
	once     sync.Once
)

func GetGlobalTTLMapInstance(gcdelay time.Duration) *GlobalTTLMap {
	once.Do(func() {
		instance = &GlobalTTLMap{
			Cache: NewTTLMap(gcdelay),
		}
	})
	return instance
}
