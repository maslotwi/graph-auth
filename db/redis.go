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
		redisClient = redis.NewClient(&redis.Options{
			Addr: environment.RedisAddr,
		})
		redisErr = redisClient.Ping(context.Background()).Err()
	})
	return redisClient, redisErr
}
