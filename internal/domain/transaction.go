package domain


import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string //custom type

const (
	StatusPending TransactionStatus = "PENDING"
	StatusFailed TransactionStatus = "FAILED"
	StatusProcessing TransactionStatus = "PROCESSING"
	StatusCategorized TransactionStatus = "CATEGORIZED"
)

type Transaction struct {
	ID uuid.UUID
	UserID uuid.UUID
	Amount float64
	Description string
	Category string
	Status TransactionStatus
	CreatedAt time.Time
}