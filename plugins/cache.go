package plugins

import (
	"time"

	"github.com/Matrix86/driplane/utils"
)

// CachePackage contains methods for use the cache (local or global)
type CachePackage struct{}

// GetCache returns the FilePackage struct
func GetCache() *CachePackage {
	return &CachePackage{}
}

// CacheResponse contains the return values
type CacheResponse struct {
	Error  error
	Status bool
	Value  string
}

// Put add a new value in the cache
func (c *CachePackage) Put(k, v string, ttl int64) {
	cache := utils.GetGlobalTTLMapInstance(5 * time.Minute).Cache
	cache.Put(k, v, ttl)
}

// Get return a cache item if it exists
func (c *CachePackage) Get(k interface{}) *CacheResponse {
	cache := utils.GetGlobalTTLMapInstance(5 * time.Minute).Cache
	v, ok := cache.Get(k)
	if ok {
		return &CacheResponse{
			Error:  nil,
			Status: true,
			Value:  v.(string),
		}
	}
	return &CacheResponse{
		Error:  nil,
		Status: false,
		Value:  "",
	}
}
