package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/storage/postgres"
)

type UserRepository struct {
	db *postgres.Postgres
}

func NewUserRepository(db *postgres.Postgres) *UserRepository {
	return &UserRepository{db}
}

func (ur *UserRepository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	const op = "repository.users.CreateUser"

	query := ur.db.Builder.Insert("users").
		Columns("username", "email", "password").
		Values(user.Username, user.Email, user.Password).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = ur.db.Pool.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return nil, fmt.Errorf("%s: %w", op, entity.ErrConflictingData)
			}
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (ur *UserRepository) Verify(ctx context.Context, userID uuid.UUID) error {
	const op = "repository.users.Verify"

	query := ur.db.Builder.Update("users").
		Set("verify", true).
		Where(squirrel.Eq{"id": userID}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = ur.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ur *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	const op = "repository.users.GetByEmail"

	var user entity.User

	query := ur.db.Builder.Select("*").
		From("users").Where(squirrel.Eq{"email": email}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = ur.db.Pool.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, entity.ErrDataNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (ur *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	const op = "repository.users.GetUserByID"

	var user entity.User

	query := ur.db.Builder.Select("*").
		From("users").
		Where(squirrel.Eq{"id": userID}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = ur.db.Pool.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, entity.ErrDataNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (ur *UserRepository) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	const op = "repository.users.UpdateUser"

	query := ur.db.Builder.Update("users").
		Set("name", squirrel.Expr("COALESCE(?, name)", user.Username)).
		Set("email", squirrel.Expr("COALESCE(?, email)", user.Email)).
		Set("password", squirrel.Expr("COALESCE(?, password)", user.Password)).
		Set("role", squirrel.Expr("COALESCE(?, role)", user.Role)).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": user.ID}).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = ur.db.Pool.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return nil, fmt.Errorf("%s: %w", op, entity.ErrConflictingData)
			}
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (ur *UserRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPass string) error {
	const op = "repository.users.UpdateUser"

	query := ur.db.Builder.Update("users").
		Set("password", newPass).
		Where(squirrel.Eq{"id": userID})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = ur.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (ur *UserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "repository.users.DeleteUser"

	query := ur.db.Builder.Delete("users").
		Where(squirrel.Eq{"id": userID})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = ur.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
