package api

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/maslotwi/graph-auth/db"
	"github.com/redis/go-redis/v9"
)

const sessionCacheTTLSeconds = 86400

func verifyTokenInGraph(token string) bool {
	if token == "" {
		return false
	}

	_, active, err := db.ActiveSessionEmail(context.Background(), token)
	return err == nil && active
}

func resolveSessionEmail(token string) (string, bool) {
	email, err := getFromRedis("session:" + token)
	if err == nil && email != "" {
		return email, true
	}

	email, active, err := db.ActiveSessionEmail(context.Background(), token)
	if err != nil || !active || email == "" {
		return "", false
	}

	_ = storeInRedis("session:"+token, email, sessionCacheTTLSeconds)
	return email, true
}

func generateSecureUUID() string {
	return uuid.NewString()
}

func storeInRedis(key string, value string, ttlSeconds int) error {
	client, err := db.Redis()
	if err != nil {
		return err
	}
	return client.Set(context.Background(), key, value, time.Duration(ttlSeconds)*time.Second).Err()
}

func getFromRedis(key string) (string, error) {
	client, err := db.Redis()
	if err != nil {
		return "", err
	}
	val, err := client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func deleteFromRedis(key string) error {
	client, err := db.Redis()
	if err != nil {
		return err
	}
	return client.Del(context.Background(), key).Err()
}

func consumeFromRedis(key string) (string, error) {
	client, err := db.Redis()
	if err != nil {
		return "", err
	}
	val, err := client.GetDel(context.Background(), key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}
