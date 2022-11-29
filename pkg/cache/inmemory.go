package cache

import (
	"errors"
	"sync"
	"time"
)

type option struct {
	ttl time.Duration
}

type Option func(*option)

func WithTTL(ttl time.Duration) Option {
	return func(o *option) {
		o.ttl = ttl
	}
}

type InMemoryCache struct {
	data    map[string]interface{}
	dataTTL map[string]time.Time

	mu sync.RWMutex
}

func NewInMemoryCache(cleanupInterval time.Duration) (*InMemoryCache, error) {
	c := &InMemoryCache{
		data:    make(map[string]interface{}),
		dataTTL: make(map[string]time.Time),
	}

	if cleanupInterval == 0 {
		return nil, errors.New("cleanupInterval cannot be 0")
	}

	go c.cleanupLoop(cleanupInterval)

	return c, nil
}

func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.data[key]
	return v, ok
}

func (c *InMemoryCache) Set(key string, value interface{}, opts ...Option) {
	c.mu.Lock()
	defer c.mu.Unlock()

	o := &option{}
	for _, opt := range opts {
		opt(o)
	}

	c.data[key] = value
	if o.ttl > 0 {
		c.dataTTL[key] = time.Now().Add(o.ttl)
	}
}

func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	delete(c.dataTTL, key)
}

func (c *InMemoryCache) deleteExpired() {
	now := time.Now()
	for key, expiration := range c.dataTTL {
		if now.After(expiration) {
			c.Delete(key)
		}
	}
}

func (c *InMemoryCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		c.deleteExpired()
	}
}
