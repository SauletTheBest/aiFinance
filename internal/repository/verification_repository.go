package repository

import (
	"context"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/google/uuid"
)

type VerificationRepository interface {
	// Create saves a newly generated code to the database
	Create(ctx context.Context, code *domain.VerificationCode) error
	
	// GetValidCode searches for a code that matches what the user typed.
	// It checks the code string, the user ID, and the type (EMAIL_VERIFY).
	GetValidCode(ctx context.Context, userID uuid.UUID, code string, codeType string) (*domain.VerificationCode, error)
	
	// MarkAsUsed updates the code so it cannot be used a second time
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
}
