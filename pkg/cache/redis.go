package cache

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(address, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, errors.New("failed to connect to Redis: " + err.Error())
	}
	return client, nil
}

func CloseRedisClient(client *redis.Client) error {
	if client == nil {
		return errors.New("client is nil")
	}
	err := client.Close()
	if err != nil {
		return errors.New("failed to close Redis client: " + err.Error())
	}
	return nil
}
