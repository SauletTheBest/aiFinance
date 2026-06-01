package repository

import (
	"context"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/google/uuid"
)

type VerificationRepository interface {
	Create(ctx context.Context, code *domain.VerificationCode) error
	GetValidCode(ctx context.Context, userID uuid.UUID, code string, codeType string) (*domain.VerificationCode, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	MarkUnusedAsUsed(ctx context.Context, userID uuid.UUID, codeType string) error
}
