package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/storage/postgres"
)

type PassHistoryRepository struct {
	db *postgres.Postgres
}

func NewPasswordHistoryRepository(db *postgres.Postgres) *PassHistoryRepository {
	return &PassHistoryRepository{db}
}

func (pr *PassHistoryRepository) GetPasswordHistory(ctx context.Context, userID uuid.UUID) ([]string, error) {
	const op = "repository.pass_history.GetPasswordHistory"

	var oldHashes []string

	query := pr.db.Builder.Select("hashed_password").
		From("user_password_history").
		Where(squirrel.Eq{"user_id": userID}).
		Limit(5)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := pr.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var hashedPassword string

		if err := rows.Scan(&hashedPassword); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		oldHashes = append(oldHashes, hashedPassword)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return oldHashes, nil
}

func (pr *PassHistoryRepository) SavePasswordToHistory(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	const op = "repository.pass_history.PasswordHistoryRepository"

	query := pr.db.Builder.Insert("user_password_history").
		Columns("user_id, hashed_password").
		Values(userID, hashedPassword)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = pr.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
