package usecase

import (
	"context"
	"fmt"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/kaspi"
	"github.com/google/uuid"
)

// ParserUseCase handles parsing bank statements and saving transactions
type ParserUseCase struct {
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepository
}

func NewParserUseCase(transactionRepo repository.TransactionRepository, userRepo repository.UserRepository) *ParserUseCase {
	return &ParserUseCase{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

// UploadStatement parses a Kaspi PDF and creates all found transactions for the user
func (uc *ParserUseCase) UploadStatement(ctx context.Context, userID uuid.UUID, filePath string) ([]*domain.Transaction, error) {
	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Очищаем старые транзакции пользователя перед импортом новых
	_, err = uc.transactionRepo.DeleteAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to clear old transactions: %w", err)
	}

	// Parse the PDF
	statement, err := kaspi.ParsePDF(filePath)
	if err != nil {
		return nil, err
	}

	// Автоматически обновляем базовый баланс пользователя в базе данных!
	if statement.StartingBalance != 0 {
		err = uc.userRepo.UpdateBaseBalance(ctx, userID, statement.StartingBalance)
		if err != nil {
			return nil, fmt.Errorf("failed to update user base balance: %w", err)
		}
	}

	// Map raw data to domain transactions and save
	transactions := make([]*domain.Transaction, 0, len(statement.Transactions))
	for _, raw := range statement.Transactions {
		tx := mapToDomain(userID, raw)

		if err := uc.transactionRepo.Create(ctx, tx); err != nil {
			return nil, err
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// mapToDomain converts a raw parsed transaction into a domain Transaction
func mapToDomain(userID uuid.UUID, raw kaspi.RawTransaction) *domain.Transaction {
	tType := domain.Expense
	amount := raw.Amount

	if amount > 0 {
		tType = domain.Income
	}

	// Store amount as positive in DB
	if amount < 0 {
		amount = -amount
	}

	return &domain.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      amount,
		Description: raw.Description,
		Category:    raw.Category,
		Type:        tType,
		Status:      domain.StatusPending,
		CreatedAt:   raw.Date,
	}
}