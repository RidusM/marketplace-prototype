package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/pkg/storage/postgres"
)

type PostgresRepository struct {
	pg *postgres.Postgres
}

func NewPostgresRepository(pg *postgres.Postgres) *PostgresRepository {
	return &PostgresRepository{pg: pg}
}

func (pr *PostgresRepository) Create(ctx context.Context, product *entity.Product) (uuid.UUID, error) {
	const op = "repository.postgres.Create"

	query, args, err := pr.pg.Builder.Insert("products").
		Columns("id", "name", "description", "price", "stock", "created_at", "updated_at").
		Values(product.Id, product.Name, product.Description, product.Price, product.Stock, product.CreatedAt, product.UpdatedAt).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	row := pr.pg.Pool.QueryRow(ctx, query, args...)
	if err = row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (pr *PostgresRepository) Get(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	const op = "repository.postgres.Get"

	query, args, err := pr.pg.Builder.Select("*").
		From("products").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	product := &entity.Product{}
	row := pr.pg.Pool.QueryRow(ctx, query, args...)
	if err = row.Scan(&product.Id, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt); err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	return product, err
}

func (pr *PostgresRepository) Update(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	const op = "repository.postgres.Update"

	query, args, err := pr.pg.Builder.Update("products").
		Set("name", product.Name).
		Set("description", product.Description).
		Set("price", product.Price).
		Set("stock", product.Stock).
		Set("updated_at", product.UpdatedAt).
		Where("id = ?", product.Id).
		ToSql()
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	_, err = pr.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	return pr.Get(ctx, product.Id)
}

func (pr *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.Delete"

	query, args, err := pr.pg.Builder.Delete("products").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = pr.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (pr *PostgresRepository) List(ctx context.Context, offset, limit int64) ([]*entity.Product, error) {
	const op = "repository.postgres.List"

	query, args, err := pr.pg.Builder.Select("id", "name", "description", "price", "stock", "created_at", "updated_at").
		From("products").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return []*entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := pr.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return []*entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	products := []*entity.Product{}
	for rows.Next() {
		product := &entity.Product{}
		if err := rows.Scan(&product.Id, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return []*entity.Product{}, fmt.Errorf("%s: %w", op, err)
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return []*entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	return products, nil
}
