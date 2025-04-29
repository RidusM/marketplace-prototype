package entity

import (
	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/utils/errs"
)

type Item struct {
	ItemID    uuid.UUID `json:"item_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     uint64    `json:"price"`
}

func NewItem(productID uuid.UUID, quantity int, price uint64) *Item {
	return &Item{ProductID: productID, Quantity: quantity, Price: price}
}

// TODO: validate in service

func ValidateItem(item *Item) (*Item, error) {
	if item.Quantity < 0 {
		return nil, errs.ErrInvalidStock
	}

	if item.Price < 0 {
		return nil, errs.ErrInvalidPrice
	}

	return item, nil
}
