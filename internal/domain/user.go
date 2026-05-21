package domain

import (
	"time"

	"github.com/google/uuid"
)


type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	Currency     string
	BaseBalance  float64   // Opening balance offset; Actual balance = BaseBalance + NetFlow
	IsVerified   bool      // 🆕 NEW FIELD: Is their email verified?
	CreatedAt    time.Time
	UpdatedAt    time.Time
}