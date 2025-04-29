package errs

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNoProfileFound = errors.New("no user found")
	ErrValidating     = errors.New("invalid data specified")
)

func HandleErrors(err error) error {
	switch {
	case errors.Is(err, ErrNoProfileFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrValidating):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
