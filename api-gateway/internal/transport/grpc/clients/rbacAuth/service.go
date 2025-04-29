package authClient

import (
	"context"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/api/rbacAuth"
)

type AuthService struct {
	auth rbacAuth.AuthServiceClient
}

func NewAuthService(auth rbacAuth.AuthServiceClient) *AuthService {
	return &AuthService{
		auth: auth,
	}
}

func (as *AuthService) SignUp(ctx context.Context, input entity.UserSignUpInput) (uuid.UUID, entity.Tokens, error) {
	_, errCreate := as.auth.SignUp(ctx, &rbacAuth.SignUpRequest{
		Username:        input.Username,
		Email:           input.Email,
		Password:        input.Password,
		PasswordConfirm: input.PasswordConfirm,
	})
	return uuid.Nil, entity.Tokens{}, errCreate
}
