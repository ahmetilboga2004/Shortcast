package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{
		client: client,
	}
}

// GetSignedURL, Redis'ten imzalı URL'i alır
func (s *RedisService) GetSignedURL(key string) (string, error) {
	ctx := context.Background()
	val, err := s.client.Get(ctx, fmt.Sprintf("signed_url:%s", key)).Result()
	if err == redis.Nil {
		return "", nil // URL bulunamadı
	}
	if err != nil {
		return "", fmt.Errorf("redis'ten URL alınırken hata: %v", err)
	}
	return val, nil
}

// SetSignedURL, imzalı URL'i Redis'e kaydeder
func (s *RedisService) SetSignedURL(key, url string, expiration time.Duration) error {
	ctx := context.Background()
	return s.client.Set(ctx, fmt.Sprintf("signed_url:%s", key), url, expiration).Err()
}

// GetMultipleSignedURLs, birden fazla imzalı URL'i Redis'ten alır
func (s *RedisService) GetMultipleSignedURLs(keys []string) (map[string]string, error) {
	ctx := context.Background()
	result := make(map[string]string)

	for _, key := range keys {
		val, err := s.client.Get(ctx, fmt.Sprintf("signed_url:%s", key)).Result()
		if err == redis.Nil {
			continue // URL bulunamadı, atla
		}
		if err != nil {
			return nil, fmt.Errorf("redis'ten URL alınırken hata: %v", err)
		}
		result[key] = val
	}

	return result, nil
}

// SetMultipleSignedURLs, birden fazla imzalı URL'i Redis'e kaydeder
func (s *RedisService) SetMultipleSignedURLs(urls map[string]string, expiration time.Duration) error {
	ctx := context.Background()
	pipe := s.client.Pipeline()

	for key, url := range urls {
		pipe.Set(ctx, fmt.Sprintf("signed_url:%s", key), url, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// DeleteSignedURL, Redis'ten imzalı URL'i siler
func (s *RedisService) DeleteSignedURL(key string) error {
	ctx := context.Background()
	return s.client.Del(ctx, fmt.Sprintf("signed_url:%s", key)).Err()
}
