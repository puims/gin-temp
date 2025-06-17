package models

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisCli *redis.Client

func init() {
	RedisCli = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisCli.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}
