package postgres

import (
	"context"
	"errors"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


type InsightRepo struct {
	db *gorm.DB
}

func NewInsightRepo(db *gorm.DB) repository.InsightRepository {
	return &InsightRepo{db: db}
}
func (r *InsightRepo) Create(ctx context.Context, insight *domain.AIInsight) error {
	return r.db.WithContext(ctx).Table("ai_insights").Create(insight).Error
}

func (r *InsightRepo) GetLatestByType(ctx context.Context, userID uuid.UUID, insightType domain.InsightType) (*domain.AIInsight, error) {
	var insight domain.AIInsight
	
	// We sort by created_at DESC to get the most recent one!
	err := r.db.WithContext(ctx).
		Table("ai_insights").
		Where("user_id = ? AND insight_type = ?", userID, insightType).
		Order("created_at DESC").
		First(&insight).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no insight exists yet
		}
		return nil, err
	}
	return &insight, nil
}	