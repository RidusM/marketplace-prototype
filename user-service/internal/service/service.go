package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"userService/internal/entity"
	"userService/internal/utils/errs"
	"userService/pkg/cache"

	"github.com/google/uuid"
)

//go:generate mockgen -source service.go -destination storage_mocks/mock.go

type Repository interface {
	Update(ctx context.Context, updatedProfile *entity.UserProfile) (uuid.UUID, error)
	Delete(ctx context.Context, profileID uuid.UUID) error
	Get(ctx context.Context, profileID uuid.UUID) (*entity.UserProfile, error)
	Create(ctx context.Context, newProfile *entity.UserProfile) (uuid.UUID, error)
}

type Cache interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
}

type Service struct {
	repo  Repository
	cache Cache
}

func New(repo Repository, cache Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) CreateProfile(ctx context.Context, userID uuid.UUID, username,
	firstname, middlename, lastname, phoneNumber, email string) (uuid.UUID, error) {
	const op = "service.createProfile"

	newProfile := entity.New(userID, username, firstname, middlename, lastname, phoneNumber, email)
	if err := entity.Valid(newProfile); err != nil {
		return uuid.Nil, errs.ErrValidating
	}

	id, err := s.repo.Create(ctx, newProfile)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Service) GetProfile(ctx context.Context, profileID uuid.UUID) (*entity.UserProfile, error) {
	const op = "service.getProfile"

	var (
		fetchedProfile *entity.UserProfile
		err            error
	)

	cmd, err := s.cache.Get(ctx, profileID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = cache.Deserialize(cmd, &fetchedProfile); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fetchedProfile, err = s.repo.Get(ctx, profileID)
	if err != nil {
		if errors.Is(err, errs.ErrNoProfileFound) {
			return nil, errs.ErrNoProfileFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	serializedProfile, err := cache.Serialize(fetchedProfile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	s.cache.Set(ctx, profileID.String(), serializedProfile, time.Hour)

	return fetchedProfile, nil
}

func (s *Service) DeleteProfile(ctx context.Context, profileID uuid.UUID) (uuid.UUID, error) {
	const op = "service.deleteProfile"

	profile, err := s.repo.Get(ctx, profileID)
	if err != nil {
		if errors.Is(err, errs.ErrNoProfileFound) {
			return uuid.Nil, errs.ErrNoProfileFound
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = s.repo.Delete(ctx, profileID); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return profile.UserID, nil
}

func (s *Service) UpdateProfile(ctx context.Context, profileID uuid.UUID, username,
	firstname, middlename, lastname, phoneNumber, email string) (uuid.UUID, error) {
	const op = "service.updateProfile"

	fetchedProfile, err := s.repo.Get(ctx, profileID)
	if err != nil {
		if errors.Is(err, errs.ErrNoProfileFound) {
			return uuid.Nil, errs.ErrNoProfileFound
		}
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	// all fields should be filled up -> if not changed -> the old value goes to creation validatingUpdated
	validatingUpdated := entity.New(uuid.New(), username, firstname, middlename, lastname, phoneNumber, email)
	if err = entity.Valid(validatingUpdated); err != nil {
		return uuid.Nil, errs.ErrValidating
	}

	fetchedProfile.UpdateFields(validatingUpdated)

	userID, err := s.repo.Update(ctx, fetchedProfile)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}
