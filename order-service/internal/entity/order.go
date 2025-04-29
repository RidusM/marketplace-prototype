package entity

import (
	"time"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/utils/errs"

	"github.com/google/uuid"
)

type Order struct {
	OrderID     uuid.UUID `json:"order_id"`
	UserID      uuid.UUID `json:"user_id"`
	TotalAmount uint64    `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewOrder(userID uuid.UUID, totalAmount uint64, status string) *Order {
	return &Order{UserID: userID, TotalAmount: totalAmount, Status: status}
}

func ValidateOrder(order *Order) (*Order, error) {
	if order.Status == "" {
		return nil, errs.ErrInvalidStatus
	}

	if order.TotalAmount < 0 {
		return nil, errs.ErrInvalidTotal
	}

	return order, nil
}
