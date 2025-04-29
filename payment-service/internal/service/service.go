package service

import (
	"context"
	"time"

	"github.com/ursulgwopp/payment-microservice/internal/entity"
)

const duration = time.Second * 30

type Database interface {
	ProcessPayment(ctx context.Context, payment entity.Payment) error
}

type Cache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type Service struct {
	db    Database
	cache Cache
}

func New(db Database, cache Cache) *Service {
	return &Service{db: db, cache: cache}
}

func (s *Service) ProcessPayment(ctx context.Context, payment *entity.Payment) (bool, error) {
	if err := s.db.ProcessPayment(ctx, entity.Payment{
		PaymentId: payment.PaymentId,
		OrderId:   payment.OrderId,
		Amount:    payment.Amount,
		Status:    payment.Status,
		CreatedAt: payment.CreatedAt,
	}); err != nil {
		return false, err
	}

	return true, nil
}
