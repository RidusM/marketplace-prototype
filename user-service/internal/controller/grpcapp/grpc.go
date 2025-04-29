package grpcapp

import (
	"context"
	"log"
	"userService/internal/entity"

	client "userService/pkg/api/client"

	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const userIdCtx = "user_id"

//go:generate mockgen -source=grpc.go -destination=service_mock.go

type Service interface {
	DeleteProfile(ctx context.Context, profileID uuid.UUID) (uuid.UUID, error)
	UpdateProfile(ctx context.Context, profileID uuid.UUID,
		username, firstname, middlename, lastname, phoneNumber, email string) (uuid.UUID, error)
	GetProfile(ctx context.Context, profileID uuid.UUID) (*entity.UserProfile, error)
	CreateProfile(ctx context.Context, userID uuid.UUID, username,
		firstname, middlename, lastname, phoneNumber, email string) (uuid.UUID, error)
}

type UserService struct {
	client.UnimplementedUserServiceServer
	service Service
}

func NewUserService(srv Service) *UserService {
	return &UserService{service: srv}
}

func (tus *UserService) CreateProfile(ctx context.Context, r *client.CreateProfileRequest) (*emptypb.Empty, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	userID, err := MapID(r.GetUserID())
	if err != nil {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "userID",
				Description: "userID trouble",
			})
	}

	if len(validationErrors) > 0 {
		for _, v := range validationErrors {
			log.Print("Validation error ", "field ", v.Field, "description ", v.Description)
		}
		stat := status.New(400, "invalid create profile request")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	if _, err = tus.service.CreateProfile(ctx, userID, r.GetUsername(), r.GetFirstname(), r.GetMiddlename(),
		r.GetLastname(), r.GetPhoneNumber(), r.GetEmail()); err != nil {
		return nil, status.Error(codes.Internal, "failed to create user profile")
	}

	return &emptypb.Empty{}, nil
}

func (tus *UserService) UpdateProfile(ctx context.Context, r *client.UpdateProfileRequest) (*emptypb.Empty, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	id, err := MapID(r.GetProfileID())
	if err != nil {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "userID",
				Description: "userID trouble",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid sign up request ")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	_, err = tus.service.UpdateProfile(ctx, id, r.GetUsername(), r.GetFirstname(), r.GetMiddlename(), r.GetLastname(), r.GetPhoneNumber(), r.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update user profile")
	}

	return &emptypb.Empty{}, nil
}

func (tus *UserService) DeleteProfile(ctx context.Context, r *client.DeleteProfileRequest) (*emptypb.Empty, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	id, err := MapID(r.GetProfileID())
	if err != nil {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "userID",
				Description: "userID trouble",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid sign up request ")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	// deleting immediately for return
	_, err = tus.service.DeleteProfile(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete user profile")
	}

	return &emptypb.Empty{}, nil
}

func (tus *UserService) GetProfile(ctx context.Context, r *client.GetProfileRequest) (*client.GetProfileResponse, error) {
	var validationErrors []*errdetails.BadRequest_FieldViolation

	id, err := MapID(r.GetProfileID())
	if err != nil {
		validationErrors = append(validationErrors,
			&errdetails.BadRequest_FieldViolation{
				Field:       "userID",
				Description: "userID trouble",
			})
	}

	if len(validationErrors) > 0 {
		stat := status.New(400, "invalid sign up request ")
		badRequest := &errdetails.BadRequest{}
		badRequest.FieldViolations = validationErrors
		s, _ := stat.WithDetails(badRequest)
		return nil, s.Err()
	}

	profile, err := tus.service.GetProfile(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user profile")
	}

	return &client.GetProfileResponse{User: MapFromEntity(profile)}, nil
}
