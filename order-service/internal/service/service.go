package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/utils/errs"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/cache"
)

const duration = time.Second * 30

const (
	StatusDelivering = "Delivery status"
	StatusShipped    = "Shipped status"
	StatusCancelled  = "Cancel status"
	StatusAccepted   = "Accepted status"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entity.Order) (uuid.UUID, error)
	GetOrder(ctx context.Context, orderID uuid.UUID) (*entity.Order, error)
	GetOrdersByUser(ctx context.Context, userID uuid.UUID) ([]entity.Order, error)
	UpdateOrder(ctx context.Context, order *entity.Order) error
	DeleteOrder(ctx context.Context, orderID uuid.UUID) error
}

type ItemRepository interface {
	AddItemOrder(ctx context.Context, item *entity.Item, orderID uuid.UUID) (uuid.UUID, error)
	DeleteItemOrder(ctx context.Context, itemID uuid.UUID) error
	UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) error
	ListItemsByOrder(ctx context.Context, orderID uuid.UUID) ([]entity.Item, error)
}

type Cache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type Service struct {
	ir    ItemRepository
	or    OrderRepository
	cache Cache
}

func New(ir ItemRepository, or OrderRepository, cache Cache) *Service {
	return &Service{ir: ir, or: or, cache: cache}
}

func (s *Service) CreateOrder(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	const op = "service.CreateOrder"
	order, err := entity.ValidateOrder(entity.NewOrder(userID, 0, StatusAccepted))
	if err != nil {
		return uuid.Nil, err
	}

	var orderID uuid.UUID
	orderID, err = s.or.CreateOrder(ctx, order)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return orderID, nil
}

func (s *Service) GetOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*entity.Order, error) {
	const op = "service.GetOrder"

	unser, err := s.cache.Get(ctx, orderID.String())
	if !errors.Is(err, errs.ErrCacheNotFound) {
		if err == nil {
			var cachedOrder entity.Order
			err = cache.Deserialize(unser, &cachedOrder)
			if err != nil {
				return nil, errs.ErrSerialization
			}
			return &cachedOrder, nil
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	order, err := s.or.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if order.UserID != userID {
		return nil, errs.ErrInvalidUserPerms
	}

	ser, err := cache.Serialize(order)
	if err != nil {
		return nil, errs.ErrSerialization
	}

	if err = s.cache.Set(ctx, orderID.String(), ser, duration); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}

func (s *Service) UpdateOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, status string, total uint64) (*entity.Order, error) {
	const op = "service.UpdateOrder"

	if status != StatusAccepted && status != StatusCancelled && status != StatusDelivering && status != StatusShipped {
		return nil, errs.ErrInvalidStatus
	}

	var order *entity.Order
	unser, err := s.cache.Get(ctx, orderID.String())
	if !errors.Is(err, errs.ErrCacheNotFound) {
		if err == nil {
			var cachedOrder entity.Order
			err = cache.Deserialize(unser, &cachedOrder)
			if err != nil {
				return nil, errs.ErrSerialization
			}
			order = &cachedOrder
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if order == nil {
		order, err = s.or.GetOrder(ctx, orderID)
		if err != nil {
			if errors.Is(err, errs.ErrOrderNotFound) {
				return nil, err
			}

			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if order.UserID != userID {
		return nil, errs.ErrInvalidUserPerms
	}

	order.Status = status
	order.TotalAmount = total

	validatedOrder, err := entity.ValidateOrder(order)
	if err != nil {
		return nil, err
	}

	err = s.or.UpdateOrder(ctx, validatedOrder)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return validatedOrder, nil
}

func (s *Service) DeleteOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error {
	const op = "service.DeleteOrder"

	order, err := s.or.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return err
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	if order.UserID != userID {
		return errs.ErrInvalidUserPerms
	}

	if err = s.or.DeleteOrder(ctx, orderID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) ListOrdersByUser(ctx context.Context, userID uuid.UUID) ([]entity.Order, error) {
	const op = "service.ListOrdersByUser"

	orders, err := s.or.GetOrdersByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return orders, nil
}

func (s *Service) AddItemOrder(ctx context.Context, orderID uuid.UUID, productID uuid.UUID, productPrice uint64, quantity int) (uuid.UUID, error) {
	const op = "service.AddItemOrder"

	item, err := entity.ValidateItem(entity.NewItem(productID, quantity, productPrice))
	if err != nil {
		return uuid.Nil, err
	}

	itemID, err := s.ir.AddItemOrder(ctx, item, orderID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return uuid.Nil, err
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	order, err := s.or.GetOrder(ctx, orderID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	order.TotalAmount += item.Price * uint64(item.Quantity)

	if err = s.or.UpdateOrder(ctx, order); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return itemID, nil
}

func (s *Service) DeleteItemOrder(ctx context.Context, itemID uuid.UUID) error {
	const op = "service.DeleteItemOrder"

	if err := s.ir.DeleteItemOrder(ctx, itemID); err != nil {
		if errors.Is(err, errs.ErrItemNotFound) {
			return err
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) error {
	const op = "service.UpdateItemOrder"

	if quantity <= 0 {
		return errs.ErrInvalidStock
	}

	if err := s.ir.UpdateItem(ctx, itemID, quantity); err != nil {
		if errors.Is(err, errs.ErrItemNotFound) {
			return err
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) ListItemsByOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) ([]entity.Item, error) {
	const op = "service.ListItemsByOrder"

	order, err := s.or.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if order.UserID != userID {
		return nil, errs.ErrInvalidUserPerms
	}

	items, err := s.ir.ListItemsByOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return items, nil
}
