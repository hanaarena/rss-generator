package cacheService

import (
	"sync"
)

// Cacher defines the interface for a cache service.
type Cacher interface {
	Get(key string) (string, bool)
	Set(key, value string)
	Delete(key string)
}

// memoryCache implements the Cacher interface.
type memoryCache struct {
	cache map[string]string
	mu    sync.RWMutex
}

// NewMemoryCache creates instance of memoryCache
func NewMemoryCache() *memoryCache {
	return &memoryCache{
		cache: make(map[string]string),
	}
}

func (c *memoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.cache[key]
	return value, ok
}

func (c *memoryCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = value
}

func (c *memoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}
