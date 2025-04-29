package grpcapp

import (
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/utils/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleErrors(err error) error {
	switch err {
	case errs.ErrInvalidStatus, errs.ErrInvalidPrice, errs.ErrInvalidStock, errs.ErrInvalidTotal, errs.ErrInvalidID:
		return status.Error(codes.InvalidArgument, err.Error())
	case errs.ErrItemNotFound, errs.ErrOrderNotFound, errs.ErrUserNotFound:
		return status.Error(codes.NotFound, err.Error())
	case errs.ErrInvalidUserPerms:
		return status.Error(codes.PermissionDenied, err.Error())
	case errs.ErrSerialization:
		return status.Error(codes.Internal, err.Error())
	case nil:
		return nil
	default:
		return status.Error(codes.Unimplemented, err.Error())
	}
}
