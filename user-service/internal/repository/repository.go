package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"userService/internal/entity"
	"userService/internal/utils/errs"
	"userService/pkg/storage/postgres"
)

const (
	profilesTable     = "profile_schema.profiles"
	userIdColumn      = "user_id"
	profileIdColumn   = "profile_id"
	usernameColumn    = "username"
	firstnameColumn   = "first_name"
	middlenameColumn  = "middle_name"
	lastnameColumn    = "last_name"
	phoneNumberColumn = "phone_number"
	emailColumn       = "email"
	//createdAtColumn   = "created_at"
	//updatedAtColumn   = "updated_at"
)

type Repository struct {
	db *postgres.Postgres
}

func New(db *postgres.Postgres) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, newProfile *entity.UserProfile) (uuid.UUID, error) {
	const op = "repository.create"

	query, args, err := r.db.Builder.Insert(profilesTable).Columns(profileIdColumn, userIdColumn, usernameColumn, firstnameColumn,
		middlenameColumn, lastnameColumn, phoneNumberColumn, emailColumn).
		Values(newProfile.ProfileID,
			newProfile.UserID,
			newProfile.Username,
			newProfile.Firstname,
			newProfile.Middlename,
			newProfile.Lastname,
			newProfile.PhoneNumber,
			newProfile.Email,
			//newProfile.CreatedAt,
			//newProfile.UpdatedAt
		).
		PlaceholderFormat(sq.Dollar).
		Suffix(fmt.Sprintf("RETURNING %s", userIdColumn)).
		ToSql()

	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	row := r.db.Pool.QueryRow(ctx, query, args...)

	var id uuid.UUID
	if err = row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *Repository) Update(ctx context.Context, updatedProfile *entity.UserProfile) (uuid.UUID, error) {
	const op = "repository.update"

	query, args, err := r.db.Builder.Update(profilesTable).
		Set(usernameColumn, updatedProfile.Username).
		Set(firstnameColumn, updatedProfile.Firstname).
		Set(middlenameColumn, updatedProfile.Middlename).
		Set(lastnameColumn, updatedProfile.Lastname).
		Set(phoneNumberColumn, updatedProfile.PhoneNumber).
		Set(emailColumn, updatedProfile.Email).
		//Set(updatedAtColumn, updatedProfile.UpdatedAt).
		Where(sq.Eq{profileIdColumn: updatedProfile.ProfileID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return updatedProfile.UserID, nil
}

func (r *Repository) Delete(ctx context.Context, profileID uuid.UUID) error {
	const op = "repository.delete"

	query, args, err := r.db.Builder.Delete(profilesTable).Where(sq.Eq{profileIdColumn: profileID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) Get(ctx context.Context, profileID uuid.UUID) (*entity.UserProfile, error) {
	const op = "repository.get"

	query, args, err := r.db.Builder.
		Select("*").From(profilesTable).Where(sq.Eq{profileIdColumn: profileID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := r.db.Pool.QueryRow(ctx, query, args...)

	var fetchedProfile entity.UserProfile
	if err = row.Scan(&fetchedProfile.ProfileID, &fetchedProfile.UserID,
		&fetchedProfile.Username,
		&fetchedProfile.Firstname,
		&fetchedProfile.Middlename,
		&fetchedProfile.Lastname,
		&fetchedProfile.PhoneNumber,
		&fetchedProfile.Email,
		&fetchedProfile.CreatedAt,
		&fetchedProfile.UpdatedAt); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNoProfileFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &fetchedProfile, nil
}
