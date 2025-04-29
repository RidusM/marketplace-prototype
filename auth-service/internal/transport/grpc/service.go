package grpcserver

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/api/rbacAuth"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	SignUp(ctx context.Context, input entity.UserSignUpInput) (uuid.UUID, entity.Tokens, error)
	Verify(ctx context.Context, userID uuid.UUID, verifyToken []byte) error
	SignIn(ctx context.Context, input entity.UserSignInInput) (entity.Tokens, error)
	RefreshSession(ctx context.Context, userRole entity.UserRole, userID, deviceID uuid.UUID) (entity.Tokens, error)
	LogOut(ctx context.Context, tokens entity.Tokens) error
	ChangePassword(ctx context.Context, input entity.UserChangePasswordInput) error
	ConfirmChangePassword(ctx context.Context, userID uuid.UUID, verifyToken []byte, newPassword string) error
	GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error)
	UpdateUser(ctx context.Context, userID *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type AuthService struct {
	rbacAuth.UnimplementedAuthServiceServer
	service Service
}

func NewAuthService(service Service) *AuthService {
	return &AuthService{
		service: service,
	}
}

func (s *AuthService) SignUp(ctx context.Context, input *rbacAuth.SignUpRequest) (*rbacAuth.SignUpResponse, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.Username == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "username",
				Description: "username cannot be null",
			})
	}

	if input.Email == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "email",
				Description: "email cannot be null",
			})
	}

	if input.Password == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "password",
				Description: "password cannot be null",
			})
	}

	if input.PasswordConfirm == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "password confirm",
				Description: "password confirm cannot be null",
			})
	}

	if input.Password != input.PasswordConfirm {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "password confirm",
				Description: "password and password confirm do not match",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid sign up request ")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	userID, tokens, err := s.service.SignUp(ctx, entity.UserSignUpInput{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		if errors.Is(err, entity.ErrConflictingData) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &rbacAuth.SignUpResponse{
		UserId: userID.String(),
		Tokens: &rbacAuth.TokensResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
	}, nil
}

func (s *AuthService) Verify(ctx context.Context, input *rbacAuth.VerifyRequest) (*emptypb.Empty, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.UserId == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "user id",
				Description: "user id cannot be null",
			})
	}

	if string(input.VerifyToken) == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "verify token",
				Description: "verify token cannot be null",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid verify request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	err := s.service.Verify(ctx, uuid.MustParse(input.UserId), input.VerifyToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to verify")
	}

	return nil, nil
}

func (s *AuthService) SignIn(ctx context.Context, input *rbacAuth.SignInRequest) (*rbacAuth.TokensResponse, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.Email == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "email",
				Description: "email cannot be null",
			})
	}

	if input.Password == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "password",
				Description: "password cannot be null",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid sign in request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	tokens, err := s.service.SignIn(ctx, entity.UserSignInInput{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		if errors.Is(err, entity.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &rbacAuth.TokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthService) RefreshSession(ctx context.Context, input *rbacAuth.RefreshSessionRequest) (*rbacAuth.TokensResponse, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.UserId == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "user id",
				Description: "user id cannot be null",
			})
	}

	if input.DeviceId == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "device id",
				Description: "device id cannot be null",
			})
	}

	if input.UserRole == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "user role",
				Description: "user role cannot be null",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid refresh session request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	tokens, err := s.service.RefreshSession(ctx, entity.UserRole(input.UserRole), uuid.MustParse(input.UserId), uuid.MustParse(input.DeviceId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to set session")
	}

	return &rbacAuth.TokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthService) LogOut(ctx context.Context, input *rbacAuth.LogOutRequest) (*emptypb.Empty, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.AccessToken == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "access token",
				Description: "access token cannot be null",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid log out request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	err := s.service.LogOut(ctx, entity.Tokens{
		AccessToken:  input.AccessToken,
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return nil, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, input *rbacAuth.ChangePasswordRequest) (*emptypb.Empty, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.Email == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "email",
				Description: "email cannot be null",
			})
	}

	if input.NewPassword == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "new password",
				Description: "new password cannot be null",
			})
	}

	if input.NewPasswordConfirm == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "new password confirm",
				Description: "new password confirm cannot be null",
			})
	}

	if input.NewPassword != input.NewPasswordConfirm {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "password confirm",
				Description: "new password and new password confirm do not match",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid change password request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	err := s.service.ChangePassword(ctx, entity.UserChangePasswordInput{
		Email:       input.Email,
		NewPassword: input.NewPassword,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	return nil, nil
}

func (s *AuthService) ConfirmChangePassword(ctx context.Context, input *rbacAuth.ConfirmChangePasswordRequest) (*emptypb.Empty, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.UserId == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "user id",
				Description: "user id cannot be null",
			})
	}

	if input.VerifyToken == nil {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "verify token",
				Description: "verify token cannot be null",
			})
	}

	if input.NewPassword == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "new password",
				Description: "new password cannot be null",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid confirm change password request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	err := s.service.ConfirmChangePassword(ctx, uuid.MustParse(input.UserId), input.VerifyToken, input.NewPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to confirm change password")
	}

	return nil, nil
}

func (s *AuthService) GetUser(ctx context.Context, input *rbacAuth.GetUserRequest) (*rbacAuth.UserResponse, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.UserId == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "user id",
				Description: "user id cannot be null",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid get user request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	user, err := s.service.GetUser(ctx, uuid.MustParse(input.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &rbacAuth.UserResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		Role:      string(user.Role),
		Verified:  user.Verified,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, input *rbacAuth.UpdateUserRequest) (*rbacAuth.UserResponse, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.UserId == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "user id",
				Description: "user id cannot be null",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid Update User request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	user, err := s.service.UpdateUser(ctx, &entity.User{
		ID:       uuid.MustParse(input.UserId),
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
		Role:     entity.UserRole(input.GetRole()),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &rbacAuth.UserResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		Role:      string(user.Role),
		Verified:  user.Verified,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, input *rbacAuth.DeleteUserRequest) (*emptypb.Empty, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	if input.UserId == "" {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "user id",
				Description: "user id cannot be null",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid delete user request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	err := s.service.DeleteUser(ctx, uuid.MustParse(input.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return nil, nil
}
