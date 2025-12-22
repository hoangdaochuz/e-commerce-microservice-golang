package cache_pkg

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalCache_Get(t *testing.T) {
	t.Run("returns nil when key not found", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		result, err := cache.Get(ctx, "nonexistent")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns value when key exists", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		err := cache.Set(ctx, "key1", "value1")
		require.NoError(t, err)

		result, err := cache.Get(ctx, "key1")

		require.NoError(t, err)
		assert.Equal(t, "value1", result)
	})

	t.Run("returns correct value for different types", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		testCases := []struct {
			key   string
			value interface{}
		}{
			{"string", "hello"},
			{"int", 42},
			{"float", 3.14},
			{"bool", true},
			{"slice", []int{1, 2, 3}},
			{"map", map[string]int{"a": 1}},
		}

		for _, tc := range testCases {
			t.Run(tc.key, func(t *testing.T) {
				err := cache.Set(ctx, tc.key, tc.value)
				require.NoError(t, err)

				result, err := cache.Get(ctx, tc.key)
				require.NoError(t, err)
				assert.Equal(t, tc.value, result)
			})
		}
	})
}

func TestLocalCache_Set(t *testing.T) {
	t.Run("sets value successfully", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		err := cache.Set(ctx, "key", "value")

		require.NoError(t, err)

		result, _ := cache.Get(ctx, "key")
		assert.Equal(t, "value", result)
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		cache.Set(ctx, "key", "value1")
		cache.Set(ctx, "key", "value2")

		result, _ := cache.Get(ctx, "key")
		assert.Equal(t, "value2", result)
	})
}

func TestLocalCache_SetEx(t *testing.T) {
	t.Run("sets value with expiration", func(t *testing.T) {
		cache := NewLocalCacheWithExpiration(time.Hour, time.Hour)
		ctx := context.Background()

		err := cache.SetEx(ctx, "expiring_key", "value", 1)

		require.NoError(t, err)

		result, _ := cache.Get(ctx, "expiring_key")
		assert.Equal(t, "value", result)
	})

	t.Run("value expires after TTL", func(t *testing.T) {
		cache := NewLocalCacheWithExpiration(100*time.Millisecond, 50*time.Millisecond)
		ctx := context.Background()

		err := cache.SetEx(ctx, "short_ttl", "value", 1)
		require.NoError(t, err)

		// Value should exist immediately
		result, _ := cache.Get(ctx, "short_ttl")
		assert.Equal(t, "value", result)

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Value should be gone (or nil)
		result, _ = cache.Get(ctx, "short_ttl")
		assert.Nil(t, result)
	})
}

func TestLocalCache_Delete(t *testing.T) {
	t.Run("deletes existing key", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		cache.Set(ctx, "to_delete", "value")
		err := cache.Delete(ctx, "to_delete")

		require.NoError(t, err)

		result, _ := cache.Get(ctx, "to_delete")
		assert.Nil(t, result)
	})

	t.Run("no error when deleting nonexistent key", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		err := cache.Delete(ctx, "nonexistent")

		require.NoError(t, err)
	})
}

func TestLocalCache_IsExist(t *testing.T) {
	t.Run("returns true when key exists", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		cache.Set(ctx, "existing", "value")
		exists, err := cache.IsExist(ctx, "existing")

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false when key does not exist", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		exists, err := cache.IsExist(ctx, "nonexistent")

		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestLocalCache_GetOrSet(t *testing.T) {
	t.Run("returns existing value without calling callback", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		cache.Set(ctx, "existing", "cached_value")
		callbackCalled := false

		result, err := cache.GetOrSet(ctx, "existing", func() (any, error) {
			callbackCalled = true
			return "new_value", nil
		})

		require.NoError(t, err)
		assert.Equal(t, "cached_value", result)
		assert.False(t, callbackCalled)
	})

	t.Run("calls callback and caches result when key not found", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		result, err := cache.GetOrSet(ctx, "new_key", func() (any, error) {
			return "computed_value", nil
		})

		require.NoError(t, err)
		assert.Equal(t, "computed_value", result)

		// Verify it was cached
		cached, _ := cache.Get(ctx, "new_key")
		assert.Equal(t, "computed_value", cached)
	})

	t.Run("returns error from callback", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()
		expectedErr := errors.New("callback error")

		result, err := cache.GetOrSet(ctx, "error_key", func() (any, error) {
			return nil, expectedErr
		})

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("prevents thundering herd with singleflight", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		var callCount int
		var mu sync.Mutex

		var wg sync.WaitGroup
		concurrency := 10

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cache.GetOrSet(ctx, "herd_key", func() (any, error) {
					mu.Lock()
					callCount++
					mu.Unlock()
					time.Sleep(50 * time.Millisecond) // Simulate slow computation
					return "value", nil
				})
			}()
		}

		wg.Wait()

		// With singleflight, callback should be called only once
		assert.Equal(t, 1, callCount)
	})
}

func TestLocalCache_GetOrSetWithEx(t *testing.T) {
	t.Run("returns existing value without calling callback", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		ctx := context.Background()

		cache.Set(ctx, "existing", "cached_value")
		callbackCalled := false

		result, err := cache.GetOrSetWithEx(ctx, "existing", func() (any, error) {
			callbackCalled = true
			return "new_value", nil
		}, 60)

		require.NoError(t, err)
		assert.Equal(t, "cached_value", result)
		assert.False(t, callbackCalled)
	})

	t.Run("calls callback and caches result with expiration", func(t *testing.T) {
		cache := NewLocalCacheWithExpiration(100*time.Millisecond, 50*time.Millisecond)
		ctx := context.Background()

		result, err := cache.GetOrSetWithEx(ctx, "expiring", func() (any, error) {
			return "computed", nil
		}, 1)

		require.NoError(t, err)
		assert.Equal(t, "computed", result)

		// Verify it was cached
		cached, _ := cache.Get(ctx, "expiring")
		assert.Equal(t, "computed", cached)
	})
}

func TestNewLocalCacheVariants(t *testing.T) {
	t.Run("NewDefaultLocalCache creates cache with default settings", func(t *testing.T) {
		cache := NewDefaultLocalCache()
		require.NotNil(t, cache)
		require.NotNil(t, cache.cache)
	})

	t.Run("NewLocalCacheNoExpiration creates cache without expiration", func(t *testing.T) {
		cache := NewLocalCacheNoExpiration()
		require.NotNil(t, cache)

		ctx := context.Background()
		cache.Set(ctx, "permanent", "value")

		// Value should persist
		result, _ := cache.Get(ctx, "permanent")
		assert.Equal(t, "value", result)
	})

	t.Run("NewLocalCacheWithExpiration creates cache with custom settings", func(t *testing.T) {
		cache := NewLocalCacheWithExpiration(time.Minute, 30*time.Second)
		require.NotNil(t, cache)
	})
}

// Verify that LocalCache implements Cache interface
func TestLocalCache_ImplementsCacheInterface(t *testing.T) {
	var _ Cache = (*LocalCache)(nil)
}
