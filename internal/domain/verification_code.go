package domain

import (
	"time"
	"github.com/google/uuid"
)

// The two types of codes we support
const (
	CodeTypeEmailVerify   = "EMAIL_VERIFY"
	CodeTypePasswordReset = "PASSWORD_RESET"
)

type VerificationCode struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Code      string    // The 4-digit string
	CodeType  string    // EMAIL_VERIFY or PASSWORD_RESET
	ExpiresAt time.Time // When does the code die?
	Used      bool      // Has it already been used?
	CreatedAt time.Time
}

