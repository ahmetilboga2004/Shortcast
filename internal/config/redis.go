package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Bağlantıyı test et
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis bağlantı hatası: %v", err)
	}

	return rdb, nil
}
