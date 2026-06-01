package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	CodeTypeEmailVerify   = "EMAIL_VERIFY"
	CodeTypePasswordReset = "PASSWORD_RESET"
)

type VerificationCode struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Code      string
	CodeType  string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}
