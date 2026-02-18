package redisclient

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedis()(*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options {
		Addr:	os.Getenv("REDIS_ADDR"),
		Password:	os.Getenv("REDIS_PASSWORD"),
		DB:	0,
		PoolSize:	50,
		MinIdleConns: 10,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test Connection
	if err := rdb.Ping(Ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}