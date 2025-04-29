package redis

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/config"
)

const (
	_defaultPoolSize    = 12000
	_defaultMinIdleCons = 200
	_defaultPoolTimeout = time.Second * 30
)

type Redis struct {
	poolSize    int
	minIdleCons int
	poolTimeout time.Duration

	Client *redis.Client
}

func New(config *config.RedisConfig, opts ...Option) (*Redis, error) {
	const op = "storage.redis.New"

	rdb := &Redis{
		poolSize:    _defaultPoolSize,
		minIdleCons: _defaultMinIdleCons,
		poolTimeout: _defaultPoolTimeout,
	}

	for _, opt := range opts {
		opt(rdb)
	}

	url := fmt.Sprintf("redis://user:%s@%s/%s",
		config.Password,
		config.Addr,
		"0",
	)

	clientConfig, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	clientConfig.PoolSize = rdb.poolSize
	clientConfig.MinIdleConns = rdb.minIdleCons
	clientConfig.PoolTimeout = rdb.poolTimeout

	rdb.Client = redis.NewClient(clientConfig)

	if err = redisotel.InstrumentTracing(rdb.Client); err != nil {
		return rdb, fmt.Errorf("%s: %w", op, err)
	}

	return rdb, nil
}

func (r *Redis) Close() error {
	if err := r.Client.Close(); err != nil {
		return fmt.Errorf("storage.redis.Close: %w", err)
	}
	return nil
}
