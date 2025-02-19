package domain

import "context"

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Has(ctx context.Context, key string) (bool, error)
}
