package redis

import "time"

type Option func(*Redis)

func PoolSize(size int) Option {
	return func(r *Redis) {
		r.poolSize = size
	}
}

func MinIdleCons(cons int) Option {
	return func(r *Redis) {
		r.minIdleCons = cons
	}
}

func PoolTimeout(timeout time.Duration) Option {
	return func(r *Redis) {
		r.poolTimeout = timeout
	}
}
