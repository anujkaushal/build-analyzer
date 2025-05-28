package cache

import (
	"sync"
	"time"
)

type Cache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	data      interface{}
	timestamp time.Time
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]*cacheItem),
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{
		data:      value,
		timestamp: time.Now(),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Cache invalidation after 1 hour
	if time.Since(item.timestamp) > time.Hour {
		delete(c.items, key)
		return nil, false
	}

	return item.data, true
}
