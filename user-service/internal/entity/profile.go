package entity

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"time"
)

type UserProfile struct {
	UserID    uuid.UUID
	ProfileID uuid.UUID
	Username  string `validate:"required"`
	Fullname
	PhoneNumber string `validate:"e164"`
	Email       string `validate:"required,email"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Fullname struct {
	Firstname  string
	Middlename string
	Lastname   string
}

func New(userID uuid.UUID, username, firstname, middlename, lastname, phoneNumber, email string) *UserProfile {
	return &UserProfile{
		UserID:      userID,
		ProfileID:   uuid.New(),
		Fullname:    Fullname{Firstname: firstname, Middlename: middlename, Lastname: lastname},
		Username:    username,
		PhoneNumber: phoneNumber,
		Email:       email,
		//CreatedAt:   time.Now(),
		//UpdatedAt:   time.Now(),
	}
}

func Valid(u *UserProfile) error {
	const op = "entity.valid"
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.StructExcept(*u, "UserID", "ProfileID", "Fullname", "CreatedAt", "UpdatedAt"); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *UserProfile) UpdateFields(forUpdating *UserProfile) {
	u.Username = forUpdating.Username
	u.Firstname = forUpdating.Firstname
	u.Middlename = forUpdating.Middlename
	u.Lastname = forUpdating.Lastname
	u.Email = forUpdating.Email
	u.PhoneNumber = forUpdating.PhoneNumber
	//u.UpdatedAt = time.Now()
}
