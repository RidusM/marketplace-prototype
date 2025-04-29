package repository

import (
	"context"
	"fmt"

	"github.com/ursulgwopp/payment-microservice/internal/config"
	"github.com/ursulgwopp/payment-microservice/internal/entity"
	"github.com/ursulgwopp/payment-microservice/pkg/storage/postgres"
)

type PostgresRepository struct {
	db *postgres.Postgres
}

func NewPostgresRepository(config *config.PostgresConfig) (*PostgresRepository, error) {
	const op = "repository.postgres.NewPostgresRepository"

	db, err := postgres.New(config)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &PostgresRepository{db: db}, nil
}

func (pr *PostgresRepository) ProcessPayment(ctx context.Context, payment entity.Payment) error {
	return nil
}
