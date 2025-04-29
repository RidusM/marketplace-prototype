package errs

import (
	"errors"
)

var (
	ErrInvalidPrice     = errors.New("invalid price")
	ErrInvalidStock     = errors.New("invalid stock")
	ErrInvalidTotal     = errors.New("invalid total amount")
	ErrOrderNotFound    = errors.New("order not found")
	ErrItemNotFound     = errors.New("order item not found")
	ErrInvalidUserPerms = errors.New("user does not have enough permission")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrInvalidID        = errors.New("invalid id")
	ErrSerialization    = errors.New("error while serialize")
	ErrCacheNotFound    = errors.New("not found value in cache")
)
