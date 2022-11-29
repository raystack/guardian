package cache_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/odpf/guardian/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryCache_ConcurrentGetSetDelete(t *testing.T) {
	c, err := cache.NewInMemoryCache(1 * time.Second)
	require.NoError(t, err)
	input := map[string]interface{}{}
	for i := 0; i < 1000; i++ {
		input[fmt.Sprintf("key%v", i)] = i
	}
	var wg sync.WaitGroup

	t.Run("setting multiple values concurrently", func(t *testing.T) {
		for k, v := range input {
			wg.Add(1)
			go func(k string, v interface{}) {
				defer wg.Done()

				c.Set(k, v)
			}(k, v)
		}
		wg.Wait()
	})

	t.Run("getting and then deleting multiple values concurrently", func(t *testing.T) {
		for k, expectedValue := range input {
			wg.Add(1)
			go func(k string, expectedValue interface{}) {
				defer wg.Done()

				actualValue, ok := c.Get(k)
				assert.True(t, ok)
				assert.Equal(t, expectedValue, actualValue)

				c.Delete(k)
			}(k, expectedValue)
		}
		wg.Wait()
	})

	t.Run("cached values should no longer exists", func(t *testing.T) {
		for k := range input {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()

				latestValue, ok := c.Get(k)
				assert.False(t, ok)
				assert.Nil(t, latestValue)
			}(k)
		}
		wg.Wait()
	})
}

func TestInMemoryCache_CacheExpiration(t *testing.T) {
	cleanupInterval := 1 * time.Second
	c, err := cache.NewInMemoryCache(cleanupInterval)
	require.NoError(t, err)

	testCases := []struct {
		key   string
		value interface{}
		ttl   time.Duration
	}{
		{"key1", "value1", 1 * time.Second},
		{"key2", "value2", 2 * time.Second},
		{"key3", "value3", 1500 * time.Millisecond},
	}

	for _, tc := range testCases {
		c.Set(tc.key, tc.value, cache.WithTTL(tc.ttl))
	}

	time.Sleep(3 * time.Second)

	for _, tc := range testCases {
		_, ok := c.Get(tc.key)
		assert.False(t, ok)
	}
}
