package usecase

import (
	"context"

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

	// Parse the PDF
	rawTransactions, err := kaspi.ParsePDF(filePath)
	if err != nil {
		return nil, err
	}

	// Map raw data to domain transactions and save
	transactions := make([]*domain.Transaction, 0, len(rawTransactions))
	for _, raw := range rawTransactions {
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
		Status:      domain.StatusCategorized,
		CreatedAt:   raw.Date,
	}
}