package entity

import (
	"errors"
)

var (
	ErrEmptyPassword      = errors.New("password cannot be empty")
	ErrDataNotFound       = errors.New("data not found")
	ErrNoUpdatedData      = errors.New("no data to update")
	ErrConflictingData    = errors.New("data conflicts with existing data in unique column")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
