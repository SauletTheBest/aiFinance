package repository

import (
	"context"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/google/uuid"
)

type GoalRepository interface {
	Create(ctx context.Context, goal *domain.Goal) error
	GetByID(ctx context.Context, id uuid.UUID)(*domain.Goal, error)
	GetByUserID(ctx context.Context, userID uuid.UUID)([]*domain.Goal, error)
	Update(ctx context.Context, goal *domain.Goal) error
	Delete(ctx context.Context, id uuid.UUID) error
}