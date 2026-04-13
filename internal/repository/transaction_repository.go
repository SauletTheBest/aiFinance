package repository

import (
	"context"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/google/uuid"
)

type TransactionRepository interface {
    Create(ctx context.Context, transaction *domain.Transaction) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Transaction, error)
    GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error)
    Update(ctx context.Context, transaction *domain.Transaction) error
    Delete(ctx context.Context, id uuid.UUID) error
}
