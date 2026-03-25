package domain


import (
	"time"

	"github.com/google/uuid"
)
type TransactionType string //
type TransactionStatus string //custom type

const (
	StatusPending TransactionStatus = "PENDING"
	StatusFailed TransactionStatus = "FAILED"
	StatusProcessing TransactionStatus = "PROCESSING"
	StatusCategorized TransactionStatus = "CATEGORIZED"
)

const (
	Income TransactionType = "income"
	Expense TransactionType = "expense"
)

type Transaction struct {
	ID uuid.UUID
	UserID uuid.UUID
	Amount float64
	Description string
	Category string
	Type TransactionType
	Status TransactionStatus
	CreatedAt time.Time
}