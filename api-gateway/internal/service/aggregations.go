package service

import (
	"context"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/entity"
	authClient "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/transport/grpc/clients/rbacAuth"
	userClient "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/transport/grpc/clients/user"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/logger"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	SignUp(ctx context.Context, input entity.UserSignUpInput) (uuid.UUID, entity.Tokens, error)
}

type UserService interface {
	GetUserProfile(ctx context.Context, userID uuid.UUID) (entity.User, error)
}

type AggregatorService struct {
	authService *authClient.AuthService
	userService *userClient.UserService
	log         *logger.Logger
}

func NewAggregatorService(authService *authClient.AuthService, userService *userClient.UserService, log *logger.Logger) *AggregatorService {
	return &AggregatorService{
		authService: authService,
		userService: userService,
		log:         log,
	}
}

func (as *AggregatorService) SignUpUser(ctx context.Context, input entity.UserSignUpInput) (map[string]interface{}, error) {
	userID, _, errCreateNewUser := as.authService.SignUp(ctx, input)
	if errCreateNewUser != nil {
		stat := status.Convert(errCreateNewUser)
		for _, detail := range stat.Details() {
			switch errType := detail.(type) {
			case *errdetails.BadRequest:
				for _, violation := range errType.GetFieldViolations() {
					as.log.Error("The field %s has invalid value. desc: %v",
						violation.GetField(), violation.GetDescription())
				}
			}
		}
		return nil, errCreateNewUser
	}

	as.log.Info("Profile created successfully")

	createUserProfileResponse, errCreateUserProfile := as.userService.CreateProfile(ctx, userID, input.Username, "", "", "", "", input.Email)
	if errCreateUserProfile != nil {
		stat := status.Convert(errCreateUserProfile)
		for _, detail := range stat.Details() {
			switch errType := detail.(type) {
			case *errdetails.BadRequest:
				for _, violation := range errType.GetFieldViolations() {
					as.log.Error("The field %s has invalid value. desc: %v",
						violation.GetField(), violation.GetDescription())
				}
			}
		}
		return nil, errCreateUserProfile
	}

	// Агрегируем данные
	return map[string]interface{}{
		"user_id":      userID,
		"user_profile": createUserProfileResponse,
	}, nil
}
