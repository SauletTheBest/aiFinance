package repository

import (
	"context"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/google/uuid"
)

type InsightRepository interface{
	Create(ctx context.Context, insight *domain.AIInsight) error
	GetLatestByType(ctx context.Context, userID uuid.UUID, insightType domain.InsightType) (*domain.AIInsight, error)
}