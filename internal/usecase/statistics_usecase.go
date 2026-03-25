package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
)

type StatisticsUsecase struct {
	statisticsRepo  repository.StatisticsRepository
	transactionRepo repository.TransactionRepository // можно оставить на будущее
	userRepo 		repository.UserRepository//
}

func NewStatisticsUsecase(
	statisticsRepo repository.StatisticsRepository,
	transactionRepo repository.TransactionRepository,
	userRepo        repository.UserRepository,
) *StatisticsUsecase {
	return &StatisticsUsecase{
		statisticsRepo:  statisticsRepo,
		transactionRepo: transactionRepo,
		userRepo: userRepo,
	}
}

// Просто прокидываем — это нормально
func (uc *StatisticsUsecase) GetBalance(ctx context.Context, userID uuid.UUID) (*domain.Balance, error) {
	balance, err := uc.statisticsRepo.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. получаем пользователя
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. добавляем валюту
	balance.Currency = user.Currency

	return balance, nil
}

// Основная бизнес-логика теперь здесь
func (uc *StatisticsUsecase) GetStatistics(
	ctx context.Context,
	userID uuid.UUID,
	periodStart, periodEnd *time.Time,
) (*domain.Statistics, error) {

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. Баланс
	balance, err := uc.statisticsRepo.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Доходы / расходы
	income, expenses, err := uc.statisticsRepo.GetIncomeExpense(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	// 3. Категории
	categories, err := uc.statisticsRepo.GetCategoryBreakdown(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	balance.Currency = user.Currency

	// 4. Собираем финальный объект
	stats := &domain.Statistics{
		Balance:           *balance,
		Income:            income,
		Expenses:          expenses, // делаем положительными
		NetFlow:           income - expenses,
		CategoryBreakdown: categories,
	}

	// 5. Период (аккуратно с nil)
	if periodStart != nil {
		stats.PeriodStart = *periodStart
	}
	if periodEnd != nil {
		stats.PeriodEnd = *periodEnd
	}

	return stats, nil
}

// Просто прокидываем (или можно удалить если не нужен)
func (uc *StatisticsUsecase) GetCategoryBreakdown(
	ctx context.Context,
	userID uuid.UUID,
	periodStart, periodEnd *time.Time,
) ([]*domain.CategoryStats, error) {
	return uc.statisticsRepo.GetCategoryBreakdown(ctx, userID, periodStart, periodEnd)
}