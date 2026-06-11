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
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	netFlow, err := uc.statisticsRepo.GetNetFlow(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Actual balance = user's opening offset + net flow from all transactions
	balance := &domain.Balance{
		UserID:    userID.String(),
		Total:     user.BaseBalance + netFlow.Total,
		Currency:  user.Currency,
		UpdatedAt: netFlow.UpdatedAt,
	}

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

	netFlow, err := uc.statisticsRepo.GetNetFlow(ctx, userID)
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
		}
	}

	actualBalance := domain.Balance{
		UserID:    userID.String(),
		Total:     user.BaseBalance + netFlow.Total,
		Currency:  user.Currency,
		UpdatedAt: netFlow.UpdatedAt,
	}

	stats := &domain.Statistics{
		Balance:           actualBalance,
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

// UpdateBalance lets the user set their real-world total balance.
// It calculates the required opening offset so that:
//
//	user.BaseBalance + netFlow == newTotalBalance
func (uc *StatisticsUsecase) UpdateBalance(ctx context.Context, userID uuid.UUID, newTotalBalance float64) error {
	// Fetch current net flow (income - expenses) across all time
	netFlow, err := uc.statisticsRepo.GetNetFlow(ctx, userID)
	if err != nil {
		return err
	}

	// Reverse-calculate the opening offset needed to hit the target balance
	newBaseBalance := newTotalBalance - netFlow.Total

	return uc.userRepo.UpdateBaseBalance(ctx, userID, newBaseBalance)
}
