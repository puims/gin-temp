package models

import (
	"context"
	"fmt"
	"gin-temp/config"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis() *redis.Client {
	redisCli := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(
			"%s:%d",
			config.Viper.GetString("redis.host"),
			config.Viper.GetInt("redis.port"),
		),
		Password: config.Viper.GetString("redis.password"),
		DB:       config.Viper.GetInt("redis.db"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisCli.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return redisCli
}
