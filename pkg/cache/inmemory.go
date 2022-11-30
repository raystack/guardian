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

type cachedValue struct {
	value interface{}
	ttl   *time.Time
}

type InMemoryCache struct {
	data map[string]cachedValue

	mu sync.RWMutex
}

func NewInMemoryCache(cleanupInterval time.Duration) (*InMemoryCache, error) {
	c := &InMemoryCache{
		data: make(map[string]cachedValue),
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
	if !ok {
		return nil, false
	}
	return v.value, true
}

func (c *InMemoryCache) Set(key string, value interface{}, opts ...Option) {
	c.mu.Lock()
	defer c.mu.Unlock()

	o := &option{}
	for _, opt := range opts {
		opt(o)
	}

	cv := cachedValue{value: value}
	if o.ttl > 0 {
		t := time.Now().Add(o.ttl)
		cv.ttl = &t
	}
	c.data[key] = cv
}

func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

func (c *InMemoryCache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, v := range c.data {
		if now.After(*v.ttl) {
			delete(c.data, k)
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
