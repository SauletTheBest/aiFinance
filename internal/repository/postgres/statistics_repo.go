package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
)

type StatisticsRepo struct {
	db *gorm.DB
}

func NewStatisticsRepo(db *gorm.DB) repository.StatisticsRepository {
	return &StatisticsRepo{db: db}
}

//
// ===== MODELS (под результаты SQL) =====
//

type balanceModel struct {
	UserID    uuid.UUID
	Total     float64
	Currency  string
	UpdatedAt time.Time
}

type incomeExpenseModel struct {
	Income   float64
	Expenses float64
}

type categoryStatsModel struct {
	Category string
	Amount   float64
	Count    int64
}

//
// ===== MAPPERS (model → domain) =====
//

func toBalanceDomain(m *balanceModel) *domain.Balance {
	return &domain.Balance{
		UserID:    m.UserID.String(),
		Total:     m.Total,
		Currency:  m.Currency,
		UpdatedAt: m.UpdatedAt,
	}
}

func toCategoryDomain(m *categoryStatsModel) *domain.CategoryStats {
	return &domain.CategoryStats{
		Category: m.Category,
		Amount:   m.Amount,
		Count:    int(m.Count),
	}
}

//
// ===== METHODS =====
//

// Баланс пользователя
func (r *StatisticsRepo) GetBalance(ctx context.Context, userID uuid.UUID) (*domain.Balance, error) {
	var result balanceModel

	err := r.db.WithContext(ctx).
		Table("transactions").
		Select(`
			COALESCE(SUM(amount), 0) as total,
			MAX(created_at) as updated_at
		`).
		Where("user_id = ?", userID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	result.UserID = userID

	return toBalanceDomain(&result), nil
}

// Доходы и расходы
func (r *StatisticsRepo) GetIncomeExpense(
	ctx context.Context,
	userID uuid.UUID,
	periodStart, periodEnd *time.Time,
) (float64, float64, error) {

	var result incomeExpenseModel

	query := r.db.WithContext(ctx).
		Table("transactions").
		Select(`
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN amount < 0 THEN amount ELSE 0 END), 0) as expenses
		`).
		Where("user_id = ?", userID)

	if periodStart != nil {
		query = query.Where("created_at >= ?", *periodStart)
	}

	if periodEnd != nil {
		query = query.Where("created_at <= ?", *periodEnd)
	}

	if err := query.Scan(&result).Error; err != nil {
		return 0, 0, err
	}

	return result.Income, result.Expenses, nil
}

// Разбивка по категориям
func (r *StatisticsRepo) GetCategoryBreakdown(
	ctx context.Context,
	userID uuid.UUID,
	periodStart, periodEnd *time.Time,
) ([]*domain.CategoryStats, error) {

	var models []categoryStatsModel

	query := r.db.WithContext(ctx).
		Table("transactions").
		Select(`
			category,
			COALESCE(SUM(amount), 0) as amount,
			COUNT(*) as count
		`).
		Where("user_id = ? AND category != ''", userID).
		Group("category")

	if periodStart != nil {
		query = query.Where("created_at >= ?", *periodStart)
	}

	if periodEnd != nil {
		query = query.Where("created_at <= ?", *periodEnd)
	}

	if err := query.Scan(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domain.CategoryStats, len(models))
	for i, m := range models {
		result[i] = toCategoryDomain(&m)
	}

	return result, nil
}