package cache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	Data      interface{}
	CreatedAt time.Time
}

type MemoryCache struct {
	store sync.Map
	ttl   time.Duration
}

// NewMemoryCache creates a new in-memory cache with specified TTL duration
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		ttl: ttl,
	}
}

// Get retrieves a key from the cache, checking for TTL expiration
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	val, ok := c.store.Load(key)
	if !ok {
		return nil, false
	}
	entry := val.(CacheEntry)
	if time.Since(entry.CreatedAt) > c.ttl {
		c.store.Delete(key)
		return nil, false
	}
	return entry.Data, true
}

// Set stores a value in the cache with the current timestamp
func (c *MemoryCache) Set(key string, data interface{}) {
	c.store.Store(key, CacheEntry{
		Data:      data,
		CreatedAt: time.Now(),
	})
}
