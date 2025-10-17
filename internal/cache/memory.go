package cache

import (
	"sync"
	"time"
)

type CacheItem struct {
	URL       string
	ExpiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*CacheItem
}

var URLCache = &MemoryCache{
	items: make(map[string]*CacheItem),
}

// Set stores a URL in cache
func (c *MemoryCache) Set(key string, url string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		URL:       url,
		ExpiresAt: time.Now().Add(duration),
	}
}

// Get retrieves a URL from cache
func (c *MemoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return "", false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		return "", false
	}

	return item.URL, true
}

// Delete removes a URL from cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// CleanExpired removes all expired items (run periodically)
func (c *MemoryCache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
		}
	}
}

// Clear removes all items from cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
}

// Size returns the number of items in cache
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
