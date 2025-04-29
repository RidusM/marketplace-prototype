package repository

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/utils/errs"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/storage/postgres"
)

const (
	OrdersTable       = "orders"
	OrderIdColumn     = "order_id"
	UserIdColumn      = "user_id"
	TotalAmountColumn = "total_amount"
	StatusColumn      = "status"
	//CreatedAtColumn = "created_at"
	//UpdatedAtColumn = "updated_at"
)

type OrderRepository struct {
	pg *postgres.Postgres
}

func NewOrderRepository(pg *postgres.Postgres) *OrderRepository {
	return &OrderRepository{pg: pg}
}

func (or *OrderRepository) CreateOrder(ctx context.Context, order *entity.Order) (uuid.UUID, error) {
	const op = "repository.createOrder"

	query, args, err := or.pg.Builder.Insert(OrdersTable).
		Columns(UserIdColumn, TotalAmountColumn, StatusColumn).
		Values(order.UserID, order.TotalAmount, order.Status).
		Suffix("RETURNING " + OrderIdColumn).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	var orderID uuid.UUID
	if err = or.pg.Pool.QueryRow(ctx, query, args...).Scan(&orderID); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return orderID, nil
}

func (or *OrderRepository) DeleteOrder(ctx context.Context, orderID uuid.UUID) error {
	const op = "repository.deleteOrder"

	query, args, err := or.pg.Builder.Delete(OrdersTable).
		Where(sq.Eq{OrderIdColumn: orderID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	tag, err := or.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrOrderNotFound
	}

	return nil
}

func (or *OrderRepository) GetOrder(ctx context.Context, orderID uuid.UUID) (*entity.Order, error) {
	const op = "repository.GetOrder"

	query, args, err := or.pg.Builder.Select("*").From(OrdersTable).
		Where(sq.Eq{OrderIdColumn: orderID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var order entity.Order
	err = or.pg.Pool.QueryRow(ctx, query, args...).Scan(&order.OrderID,
		&order.UserID,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrOrderNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &order, nil
}

func (or *OrderRepository) GetOrdersByUser(ctx context.Context, userID uuid.UUID) ([]entity.Order, error) {
	const op = "repository.GetOrdersByUser"

	query, args, err := or.pg.Builder.Select("*").From(OrdersTable).
		Where(sq.Eq{UserIdColumn: userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var orders []entity.Order

	rows, err := or.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var order entity.Order

		err = rows.Scan(&order.OrderID, &order.UserID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return orders, nil
}

func (or *OrderRepository) UpdateOrder(ctx context.Context, order *entity.Order) error {
	const op = "repository.updateOrder"

	query, args, err := or.pg.Builder.Update(OrdersTable).
		Where(sq.Eq{OrderIdColumn: order.OrderID}).
		Set(StatusColumn, order.Status).
		Set(TotalAmountColumn, order.TotalAmount).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	tag, err := or.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrOrderNotFound
	}

	return nil
}
