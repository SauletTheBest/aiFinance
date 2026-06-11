package repository

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
)

type StatisticsRepository interface {
	GetNetFlow(ctx context.Context, userID uuid.UUID) (*domain.Balance, error)
    GetIncomeExpense(ctx context.Context, userID uuid.UUID, periodStart, periodEnd *time.Time) (float64, float64, error)
	GetCategories(ctx context.Context, userID uuid.UUID, periodStart, periodEnd *time.Time) ([]*domain.CategoryStats, error)
}
