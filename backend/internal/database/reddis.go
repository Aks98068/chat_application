package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(
	redisURL string,
) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse Redis URL: %w",
			err,
		)
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf(
			"failed to connect to Redis: %w",
			err,
		)
	}

	return client, nil
}
