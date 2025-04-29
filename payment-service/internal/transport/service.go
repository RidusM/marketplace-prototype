package grpcServer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ursulgwopp/payment-microservice/internal/entity"
	"github.com/ursulgwopp/payment-microservice/pkg/api/payment"
)

type Service interface {
	ProcessPayment(ctx context.Context, payment *entity.Payment) (bool, error)
}

type PaymentService struct {
	payment.UnsafePaymentServiceServer
	service Service
}

func NewPaymentService(service Service) *PaymentService {
	return &PaymentService{service: service}
}

func (t *PaymentService) ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error) {
	res, err := t.service.ProcessPayment(ctx, &entity.Payment{
		PaymentId: uuid.NewString(),
		OrderId:   req.OrderId,
		Amount:    req.Amount,
		Status:    req.Status,
		CreatedAt: time.Now(),
	})
	return &payment.ProcessPaymentResponse{Success: res}, err
}
