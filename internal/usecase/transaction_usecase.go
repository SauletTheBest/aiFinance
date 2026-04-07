package usecase


import (
	"context"
	"github.com/google/uuid"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"time"
)


type TransactionUsecase struct {
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepository
}

func NewTransactionUsecase(transactionRepo repository.TransactionRepository, userRepo repository.UserRepository) *TransactionUsecase {
	return &TransactionUsecase{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

// CreateTransaction - Create a new transaction for a user
func (uc *TransactionUsecase) CreateTransaction(ctx context.Context, userID uuid.UUID, amount float64, description string, category string, transactionType string, createdAt time.Time) (*domain.Transaction, error) {
	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	transaction := &domain.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      amount,
		Description: description,
		Category:    category,  // " " Will be set by AI service but i maked manually maybe will change to ""
		Type:  		 domain.TransactionType(transactionType),
		Status:      domain.StatusPending,
		CreatedAt:   createdAt,
	}

	err = uc.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetTransaction - Get a specific transaction by ID
func (uc *TransactionUsecase) GetTransaction(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	return uc.transactionRepo.GetByID(ctx, id)
}

// GetUserTransactions - Get all transactions for a user
func (uc *TransactionUsecase) GetUserTransactions(ctx context.Context, userID uuid.UUID) ([]*domain.Transaction, error) {
	return uc.transactionRepo.GetByUserID(ctx, userID)
}

// UpdateTransaction - Update transaction details
func (uc *TransactionUsecase) UpdateTransaction(ctx context.Context, id uuid.UUID, amount float64, description string, category string, createdAt time.Time) error {
	transaction, err := uc.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	transaction.Amount = amount
	transaction.Description = description
	transaction.Category = category

	if !createdAt.IsZero() {
        transaction.CreatedAt = createdAt
    }

	return uc.transactionRepo.Update(ctx, transaction)
}

// DeleteTransaction - Delete a transaction
func (uc *TransactionUsecase) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	return uc.transactionRepo.Delete(ctx, id)
}
