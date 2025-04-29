package entity

import "time"

type Payment struct {
	PaymentId     string    `json:"payment_id"`
	OrderId       string    `json:"order_id"`
	Amount        int64     `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
