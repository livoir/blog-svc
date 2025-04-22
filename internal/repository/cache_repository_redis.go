package repository

import (
	"context"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CacheRepositoryRedis struct {
	Client *redis.Client
}

func NewCacheRepositoryRedis(client *redis.Client) (domain.CacheRepository, error) {
	if client == nil {
		logger.Log.Error("Redis client is nil")
		return nil, common.ErrInternalServerError
	}
	return &CacheRepositoryRedis{Client: client}, nil
}

func (c *CacheRepositoryRedis) Clear(ctx context.Context) error {
	err := c.Client.FlushAll(ctx).Err()
	if err != nil {
		logger.Log.Error("Failed to clear cache: ", zap.Error(err))
		return common.ErrInternalServerError
	}
	return nil
}

func (c *CacheRepositoryRedis) Delete(ctx context.Context, key string) error {
	err := c.Client.Del(ctx, key).Err()
	if err != nil {
		logger.Log.Error("Failed to delete key from cache: ", zap.String("key", key), zap.Error(err))
		return common.ErrInternalServerError
	}
	return nil
}

func (c *CacheRepositoryRedis) Get(ctx context.Context, key string) (interface{}, error) {
	result, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		logger.Log.Error("Failed to get value from cache: ", zap.String("key", key), zap.Error(err))
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (c *CacheRepositoryRedis) Has(ctx context.Context, key string) (bool, error) {
	val, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		logger.Log.Error("Failed to check existence of key in cache: ", zap.String("key", key), zap.Error(err))
		return false, common.ErrInternalServerError
	}
	return val > 0, nil
}

func (c *CacheRepositoryRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		logger.Log.Error("Failed to set value in cache: ", zap.String("key", key), zap.Error(err))
		return common.ErrInternalServerError
	}
	return nil
}
