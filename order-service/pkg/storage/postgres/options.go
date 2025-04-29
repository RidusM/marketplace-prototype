package postgres

import "time"

type Option func(*Postgres)

func MaxPoolSize(size int32) Option {
	return func(p *Postgres) {
		p.maxPoolSize = size
	}
}

func MaxConnAttempts(attempts int) Option {
	return func(p *Postgres) {
		p.connAttempts = attempts
	}
}

func ConnTimeout(timeout time.Duration) Option {
	return func(p *Postgres) {
		p.connTimeout = timeout
	}
}
