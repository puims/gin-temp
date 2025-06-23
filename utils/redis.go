package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func setupRedis() *redis.Client {
	redisCli := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(
			"%s:%d",
			Viper.GetString("redis.host"),
			Viper.GetInt("redis.port"),
		),
		Password: Viper.GetString("redis.password"),
		DB:       Viper.GetInt("redis.db"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisCli.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return redisCli
}
