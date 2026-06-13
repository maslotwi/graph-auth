package db

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// Close releases database connections opened by Neo4j() and Redis().
func Close() error {
	var errs []error

	if neo4jDriver != nil {
		if err := neo4jDriver.Close(context.Background()); err != nil {
			errs = append(errs, err)
		}
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil && !errors.Is(err, redis.ErrClosed) {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
