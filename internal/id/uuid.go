package id

import (
	"github.com/google/uuid"
)

// NewV7 returns a UUID string. TODO: Implement true UUIDv7 per ADR-001.
func NewV7() string {
	// Placeholder: use v4 until v7 is implemented or available.
	return uuid.NewString()
}
