package grpcapp

import (
	"fmt"
	"userService/internal/entity"
	client "userService/pkg/api/client"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapID(id string) (uuid.UUID, error) {
	const op = "grpcapp.MapID"

	mappedID, err := uuid.ParseBytes([]byte(id))
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return mappedID, nil
}

func MapToEntity(req *client.Profile) (*entity.UserProfile, error) {
	const op = "grpcapp.MapToEntity"
	profileID, err := MapID(req.GetProfileID())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := MapID(req.GetUserID())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.UserProfile{UserID: userID,
			ProfileID:   profileID,
			Username:    req.GetUsername(),
			Fullname:    entity.Fullname{Firstname: req.GetFirstname(), Middlename: req.GetMiddlename(), Lastname: req.GetLastname()},
			PhoneNumber: req.GetPhoneNumber(),
			Email:       req.GetEmail(),
			CreatedAt:   req.GetCreatedAt().AsTime(),
			UpdatedAt:   req.GetUpdatedAt().AsTime()},
		nil
}

func MapFromEntity(profile *entity.UserProfile) *client.Profile {
	return &client.Profile{UserID: profile.UserID.String(),
		ProfileID:   profile.ProfileID.String(),
		Username:    profile.Username,
		Firstname:   profile.Firstname,
		Middlename:  profile.Middlename,
		Lastname:    profile.Lastname,
		PhoneNumber: profile.PhoneNumber,
		Email:       profile.Email,
		CreatedAt:   timestamppb.New(profile.CreatedAt),
		UpdatedAt:   timestamppb.New(profile.UpdatedAt),
	}
}
