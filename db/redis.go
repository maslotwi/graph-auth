package db

import (
	"context"
	"sync"

	"github.com/maslotwi/graph-auth/helpers/environment"
	"github.com/redis/go-redis/v9"
)

var (
	redisOnce   sync.Once
	redisClient *redis.Client
	redisErr    error
)

// Redis returns the shared Redis client. The underlying connection pool is safe
// for concurrent use across goroutines.
func Redis() (*redis.Client, error) {
	redisOnce.Do(func() {
		opts, err := redis.ParseURL(environment.RedisURL)
		if err != nil {
			redisErr = err
			return
		}

		redisClient = redis.NewClient(opts)
		redisErr = redisClient.Ping(context.Background()).Err()
	})
	return redisClient, redisErr
}
