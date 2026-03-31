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
	transactionRepo repository.TransactionRepository 
	userRepo 		repository.UserRepository
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

func (uc *StatisticsUsecase) GetBalance(ctx context.Context, userID uuid.UUID) (*domain.Balance, error) {
	balance, err := uc.statisticsRepo.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	balance.Currency = user.Currency

	return balance, nil
}

func (uc *StatisticsUsecase) GetStatistics(
	ctx context.Context,
	userID uuid.UUID,
	periodStart, periodEnd *time.Time,
) (*domain.Statistics, error) {

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	balance, err := uc.statisticsRepo.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	income, expenses, err := uc.statisticsRepo.GetIncomeExpense(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	categories, err := uc.statisticsRepo.GetCategories(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	var expenseCategories []*domain.CategoryStats
	var incomeCategories []*domain.CategoryStats

	for _, category := range categories {
		switch category.Type {
		case "expense":
			expenseCategories = append(expenseCategories, category)
		case "income":
			incomeCategories = append(incomeCategories, category)
		default:
			//i can there throw error in future
		}

	}

	balance.Currency = user.Currency
	stats := &domain.Statistics{
		Balance:           *balance,
		Income:            income,
		Expenses:          expenses,
		NetFlow:           income - expenses,
		ExpenseCategories: expenseCategories,
		IncomeCategories:  incomeCategories,  
	}

	if periodStart != nil {
		stats.PeriodStart = *periodStart
	}
	if periodEnd != nil {
		stats.PeriodEnd = *periodEnd
	}

	return stats, nil
}
