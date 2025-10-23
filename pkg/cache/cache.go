package cache_pkg

import "context"

type Cache interface {
	// Get value from cache
	Get(ctx context.Context, key string) (any, error)
	// Set value to cache
	Set(ctx context.Context, key string, value any) error
	// Set value to cache with expiration time
	SetEx(ctx context.Context, key string, value any, seconds int) error
	// Delete value from cache
	Delete(ctx context.Context, key string) error
	// Check if value exists in cache
	IsExist(ctx context.Context, key string) (bool, error)
	// Get if exist and set into cache if not
	GetOrSet(ctx context.Context, key string, callback func() (any, error)) (any, error)
	// Get if exist and set into cache if not with expiration time
	GetOrSetWithEx(ctx context.Context, key string, callback func() (any, error), seconds int) (any, error)
}
