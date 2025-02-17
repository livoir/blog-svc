package repository

import (
	"context"
	"livoir-blog/internal/domain"

	"github.com/redis/go-redis/v9"
)

type CacheRepositoryRedis struct {
	Client *redis.Client
}

func NewCacheRepositoryRedis(client *redis.Client) domain.CacheRepository {
	return &CacheRepositoryRedis{Client: client}
}

func (c *CacheRepositoryRedis) Clear(ctx context.Context) error {
	return c.Client.FlushAll(ctx).Err()
}

func (c *CacheRepositoryRedis) Delete(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

func (c *CacheRepositoryRedis) Get(ctx context.Context, key string) (interface{}, error) {
	return c.Client.Get(ctx, key).Result()
}

func (c *CacheRepositoryRedis) Has(ctx context.Context, key string) (bool, error) {
	val, err := c.Client.Exists(ctx, key).Result()
	return val > 0, err
}

func (c *CacheRepositoryRedis) Set(ctx context.Context, key string, value interface{}) error {
	return c.Client.Set(ctx, key, value, 0).Err()
}
