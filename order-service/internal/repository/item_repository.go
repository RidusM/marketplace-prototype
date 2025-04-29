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
	JunctionTable   = "orders_items"
	ItemTable       = "items"
	ItemIdColumn    = "item_id"
	ProductIdColumn = "product_id"
	QuantityColumn  = "quantity"
	PriceColumn     = "price"
)

type ItemRepository struct {
	pg *postgres.Postgres
}

func NewItemRepository(pg *postgres.Postgres) *ItemRepository {
	return &ItemRepository{pg: pg}
}

func (ir *ItemRepository) AddItemOrder(ctx context.Context, item *entity.Item, orderID uuid.UUID) (uuid.UUID, error) {
	const op = "item_repository.AddItem"

	tx, err := ir.pg.Pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	queryItems, argsItems, err := ir.pg.Builder.Insert(ItemTable).
		Columns(ProductIdColumn, QuantityColumn, PriceColumn).
		Values(item.ProductID, item.Quantity, item.Price).
		Suffix("RETURNING " + ItemIdColumn).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	var itemID uuid.UUID
	if err = tx.QueryRow(ctx, queryItems, argsItems...).Scan(&itemID); err != nil {
		defer tx.Rollback(ctx)

		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, errs.ErrOrderNotFound
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	queryJunc, argsJunc, err := ir.pg.Builder.Insert(JunctionTable).
		Columns(OrderIdColumn, ItemIdColumn).
		Values(orderID, itemID).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err = tx.Exec(ctx, queryJunc, argsJunc...); err != nil {
		tx.Rollback(ctx)
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return itemID, nil
}

func (ir *ItemRepository) DeleteItemOrder(ctx context.Context, itemID uuid.UUID) error {
	const op = "item_repository.DeleteItem"

	query, args, err := ir.pg.Builder.Delete(ItemTable).Where(sq.Eq{ItemIdColumn: itemID}).ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	tag, err := ir.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrItemNotFound
	}

	return nil
}

func (ir *ItemRepository) UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) error {
	const op = "item_repository.UpdateItem"

	query, args, err := ir.pg.Builder.Update(ItemTable).Where(sq.Eq{ItemIdColumn: itemID}).
		Set(QuantityColumn, quantity).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	tag, err := ir.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrItemNotFound
	}

	return nil
}

func (ir *ItemRepository) ListItemsByOrder(ctx context.Context, orderID uuid.UUID) ([]entity.Item, error) {
	const op = "item_repository.ListItemByOrder"

	query, args, err := ir.pg.Builder.Select(ItemIdColumn, ProductIdColumn, QuantityColumn, PriceColumn).
		From(fmt.Sprintf("%s %s", ItemTable, "i")).InnerJoin(fmt.Sprintf("%s o ON i.%s = o.%s", JunctionTable, ItemIdColumn, ItemIdColumn)).
		Where(sq.Eq{OrderIdColumn: orderID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := ir.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var itemsOrder []entity.Item
	for rows.Next() {
		var item entity.Item
		if err = rows.Scan(&item.ItemID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		itemsOrder = append(itemsOrder, item)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return itemsOrder, nil
}
