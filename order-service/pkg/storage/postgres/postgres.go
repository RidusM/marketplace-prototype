package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/exaring/otelpgx"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/config"
)

const (
	_defaultMaxPoolSize     = 1
	_defaultConnAttempts    = 10
	_defaultConnTimeout     = time.Second
	_defaultMigrateAttempts = 5
	_defaultMigrateTimeout  = time.Second
)

type Postgres struct {
	connAttempts int
	connTimeout  time.Duration
	maxPoolSize  int32

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

func New(config *config.PostgresConfig, opts ...Option) (*Postgres, error) {
	const op = "storage.postgres.New"

	url := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		config.Connection,
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)

	pg := &Postgres{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
		maxPoolSize:  _defaultMaxPoolSize,
	}

	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	poolConfig.MaxConns = pg.maxPoolSize

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = otelpgx.RecordStats(pg.Pool); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	initMigrate(url)

	return pg, nil
}

func initMigrate(dbURL string) {
	var (
		attempts = _defaultMigrateAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", dbURL)
		if err == nil {
			break
		}

		log.Println(err)
		log.Printf("Migrate: postgres is trying to connect, attempts left: %d", attempts)
		time.Sleep(_defaultMigrateTimeout)
		attempts--
	}

	if err != nil {
		log.Fatalf("Migrate: postgres connect error: %s", err)
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Migrate: up error: %s", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	log.Printf("Migrate: up success")
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
