package entity

import "github.com/google/uuid"

// GenerateID generates a unique ID that can be used as an identifier for an entity.
func GenerateID() uuid.UUID {
	return uuid.New()
}
