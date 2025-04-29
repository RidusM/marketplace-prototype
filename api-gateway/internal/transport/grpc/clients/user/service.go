package userClient

import (
	"context"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/api/client"
)

type UserService struct {
	client client.UserServiceClient
}

func NewUserService(client client.UserServiceClient) *UserService {
	return &UserService{
		client: client,
	}
}

func (us *UserService) CreateProfile(ctx context.Context, userID uuid.UUID, username, firstname, middlename, lastname, phoneNumber, email string) (uuid.UUID, error) {
	_, errCreate := us.client.CreateProfile(ctx, &client.CreateProfileRequest{
		UserID:      userID.String(),
		Username:    username,
		Firstname:   firstname,
		Middlename:  middlename,
		Lastname:    lastname,
		PhoneNumber: phoneNumber,
		Email:       email,
	})
	return uuid.Nil, errCreate
}
