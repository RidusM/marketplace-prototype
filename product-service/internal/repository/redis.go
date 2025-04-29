package repository

import (
	"context"
	"fmt"
	"time"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/pkg/storage/redis"
)

type RedisCache struct {
	cache *redis.Redis
}

func NewRedisCache(cache *redis.Redis) *RedisCache {
	return &RedisCache{cache: cache}
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	const op = "repository.redis.Set"

	err := r.cache.Client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	const op = "repository.redis.Get"

	res, err := r.cache.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	bytes := []byte(res)

	return bytes, err
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	const op = "repository.redis.Delete"

	err := r.cache.Client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
