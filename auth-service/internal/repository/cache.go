package repository

import (
	"context"
	"fmt"
	"time"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/storage/redis"
)

type CacheRepository struct {
	rdb *redis.Redis
}

func NewCacheRepository(rdb *redis.Redis) *CacheRepository {
	return &CacheRepository{rdb}
}

func (r *CacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := r.rdb.Client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("repository.cache.Set: %w", err)
	}
	return nil
}

func (r *CacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	res, err := r.rdb.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("repository.cache.Get: %w", err)
	}

	bytes := []byte(res)

	return bytes, nil
}

func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	if err := r.rdb.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("repository.cache.DeleteByPattern: %w", err)
	}

	return nil
}

func (r *CacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	const op = "repository.cache.DeleteByPattern"

	keys, err := r.rdb.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if len(keys) == 0 {
		return nil
	}

	_, err = r.rdb.Client.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
