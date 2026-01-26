package middleware

import (
	"sync"
	"time"
)

// CacheEntry represents a cached value with expiration.
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
}

// IsExpired returns true if the cache entry has expired.
func (e CacheEntry) IsExpired() bool {
	return time.Now().After(e.Expiration)
}

// Cache provides in-memory caching with TTL support.
type Cache struct {
	defaultTTL time.Duration
	entries    map[string]CacheEntry
	mu         sync.RWMutex

	// Cleanup configuration
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewCache creates a new cache with the specified default TTL.
func NewCache(defaultTTL time.Duration) *Cache {
	c := &Cache{
		defaultTTL:      defaultTTL,
		entries:         make(map[string]CacheEntry),
		cleanupInterval: defaultTTL / 2,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup goroutine
	go c.cleanupLoop()

	return c
}

// Get retrieves a value from the cache.
// Returns the value and true if found and not expired, otherwise nil and false.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		// Lazy deletion
		c.Delete(key)
		return nil, false
	}

	return entry.Value, true
}

// GetTyped retrieves a typed value from the cache.
func GetTyped[T any](c *Cache, key string) (T, bool) {
	var zero T
	value, ok := c.Get(key)
	if !ok {
		return zero, false
	}

	typed, ok := value.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// Set stores a value in the cache with the default TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	c.entries[key] = CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// Delete removes a value from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Clear removes all values from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]CacheEntry)
	c.mu.Unlock()
}

// Size returns the number of entries in the cache (including expired ones).
func (c *Cache) Size() int {
	c.mu.RLock()
	size := len(c.entries)
	c.mu.RUnlock()
	return size
}

// Keys returns all keys in the cache (including expired ones).
func (c *Cache) Keys() []string {
	c.mu.RLock()
	keys := make([]string, 0, len(c.entries))
	for k := range c.entries {
		keys = append(keys, k)
	}
	c.mu.RUnlock()
	return keys
}

// cleanupLoop periodically removes expired entries.
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup removes all expired entries.
func (c *Cache) cleanup() {
	now := time.Now()
	c.mu.Lock()
	for key, entry := range c.entries {
		if now.After(entry.Expiration) {
			delete(c.entries, key)
		}
	}
	c.mu.Unlock()
}

// Close stops the background cleanup goroutine.
func (c *Cache) Close() {
	close(c.stopCleanup)
}

// GetOrSet returns the cached value if it exists, otherwise calls the function
// to compute the value, stores it in the cache, and returns it.
func (c *Cache) GetOrSet(key string, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, ok := c.Get(key); ok {
		return value, nil
	}

	// Compute the value
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.Set(key, value)

	return value, nil
}

// GetOrSetTyped is a typed version of GetOrSet.
func GetOrSetTyped[T any](c *Cache, key string, fn func() (T, error)) (T, error) {
	var zero T

	// Try to get from cache first
	if value, ok := GetTyped[T](c, key); ok {
		return value, nil
	}

	// Compute the value
	value, err := fn()
	if err != nil {
		return zero, err
	}

	// Store in cache
	c.Set(key, value)

	return value, nil
}
