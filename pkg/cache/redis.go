package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(address, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil
	}
	return client
}
