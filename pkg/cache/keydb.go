package cache

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

func NewKeyDBClient(ctx context.Context, address, username, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
		Username: username,
	})
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, errors.New("failed to connect to KeyDB: " + err.Error())
	}
	return client, nil
}

func CloseKeyDBClient(client *redis.Client) error {
	if client == nil {
		return errors.New("client is nil")
	}
	err := client.Close()
	if err != nil {
		return errors.New("failed to close KeyDB client: " + err.Error())
	}
	return nil
}
