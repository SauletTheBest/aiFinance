package domain

import (
	"time"

	"github.com/google/uuid"
)


type User struct {
	ID uuid.UUID
	Name string
	Email string
	PasswordHash string
	Currency string
	CreatedAt time.Time
	UpdatedAt time.Time
}