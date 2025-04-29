package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/entity"
)

type (
	Manager struct {
		signingKey string
	}

	CustomClaims struct {
		DeviceID uuid.UUID
		Role     entity.UserRole

		jwt.RegisteredClaims
	}

	ParsedTokenData struct {
		UserID   uuid.UUID       `json:"user_id"`
		DeviceID uuid.UUID       `json:"device_id"`
		Role     entity.UserRole `json:"role"`
	}
)

type TokenManager interface {
	NewJWT(userID, deviceID uuid.UUID, role entity.UserRole, ttl time.Duration) (string, error)
	Parse(accessToken string) (*ParsedTokenData, error)
	NewRefreshToken() (string, error)
}

func New(signingKey string) (*Manager, error) {
	const op = "auth.manager.New"

	if signingKey == "" {
		return nil, fmt.Errorf("%s: empty signingKey", op)
	}

	return &Manager{signingKey: signingKey}, nil
}

func (m *Manager) NewJWT(userID, deviceID uuid.UUID, role entity.UserRole, ttl time.Duration) (string, error) {
	claims := CustomClaims{
		DeviceID: deviceID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	res, err := token.SignedString([]byte(m.signingKey))
	if err != nil {
		return "", fmt.Errorf("auth.manager.NewJWT: %w", err)
	}

	return res, nil
}

func (m *Manager) Parse(accessToken string) (*ParsedTokenData, error) {
	const op = "auth.manager.Parse"

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: unexpected signing method", op)
		}

		return []byte(m.signingKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	parsedData := &ParsedTokenData{
		UserID:   uuid.MustParse(claims.Subject),
		Role:     claims.Role,
		DeviceID: claims.DeviceID,
	}

	return parsedData, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	const op = "auth.manager.NewRefreshToken"

	size := 32
	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return hex.EncodeToString(b), nil
}
