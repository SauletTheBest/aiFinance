package postgres


import (
	"gorm.io/gorm"
	"time"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/google/uuid"
	"context"
)

type TransactionRepo struct {
	db *gorm.DB
}

type Transaction struct {
	ID 			uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID 		uuid.UUID `gorm:"type:uuid;index"`
	Amount 		float64
	Description string
	Category 	string
	Type		domain.TransactionType `gorm:"type:varchar(10)"`
	Status 		domain.TransactionStatus
	CreatedAt 	time.Time
}

func transactionToDomain(model *Transaction) *domain.Transaction {
	return &domain.Transaction{
		ID: 			model.ID,
		UserID: 		model.UserID,
		Amount: 		model.Amount,
		Description:	model.Description,
		Category: 		model.Category,
		Type:			model.Type,
		Status: 		model.Status,
		CreatedAt: 		model.CreatedAt,
	}
}

func transactionToModel(transaction *domain.Transaction) *Transaction {
	return &Transaction{
		ID: 			transaction.ID,
		UserID: 		transaction.UserID,
		Amount: 		transaction.Amount,
		Description:	transaction.Description,
		Category: 		transaction.Category,
		Type:			transaction.Type,
		Status: 		transaction.Status,
		CreatedAt: 		transaction.CreatedAt,
	}
}

func NewTransactionRepo(db *gorm.DB) repository.TransactionRepository {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) Create(ctx context.Context, transaction *domain.Transaction) error {
	model := transactionToModel(transaction)
	return r.db.WithContext(ctx).Create(model).Error
}



func (r *TransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	var model Transaction

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error

	if err != nil {
		return nil, err
	}
	return transactionToDomain(&model), nil
}

func (r *TransactionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Transaction, error) {
    var models []Transaction
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Find(&models).Error
    if err != nil {
        return nil, err
    }
    
    transactions := make([]*domain.Transaction, len(models))
    for i, model := range models {
        transactions[i] = transactionToDomain(&model)
    }
    return transactions, nil
}
func (r *TransactionRepo) Update(ctx context.Context, transaction *domain.Transaction) error{
	model := transactionToModel(transaction)
    return r.db.WithContext(ctx).Save(model).Error
}
func (r *TransactionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Transaction{}).Error
}

// GetByStatus returns transactions filtered by status, limited to `limit` rows.
// Used by the background worker to grab PENDING transactions for AI categorization.
func (r *TransactionRepo) GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error) {
	var models []Transaction
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	transactions := make([]*domain.Transaction, len(models))
	for i, model := range models {
		transactions[i] = transactionToDomain(&model)
	}
	return transactions, nil
}

func (r *TransactionRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	db := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&Transaction{})
	return db.RowsAffected, db.Error
}