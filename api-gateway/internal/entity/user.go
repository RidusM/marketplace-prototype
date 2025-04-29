package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	Admin  UserRole = "admin"
	Client UserRole = "client"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Password  string
	Role      UserRole
	Verified  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserSignUpInput struct {
	Username        string
	Email           string
	Password        string
	PasswordConfirm string
}
