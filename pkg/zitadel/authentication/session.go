package zitadel_authentication

import "context"

type SessionHandler[T any] interface {
	Set(ctx context.Context, key string, value T, seconds int) error
	Get(ctx context.Context, key string) (*T, error)
	Del(ctx context.Context, key string) error
}
