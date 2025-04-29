package entity

import "errors"

var (
	ErrInvalidName        = errors.New("invalid name")
	ErrInvalidDescription = errors.New("invalid description")
	ErrInvalidPrice       = errors.New("invalid price")
	ErrInvalidStock       = errors.New("invalid stock")
)
