package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Password(password string) (string, error) {
	const op = "hash.Password"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	return string(hashedPassword), nil
}

func ComparePassword(password, hashedPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("hash.ComparePassword: %w", err)
	}

	return nil
}
