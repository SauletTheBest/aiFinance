package dto


import (
	"github.com/google/uuid"
	"time"
)

type CreateTransactionRequest struct {
	Amount      float64   `json:"amount" validate:"required"`
	Description string    `json:"description"`
	Category 	string    `json:"category"`
	Type 		string 	  `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateTransactionRequest struct {
	Amount      float64 	`json:"amount,omitempty" validate:"omitempty,gt=0"`
	Description string  	`json:"description,omitempty" validate:"omitempty,max=500"`
	Category 	string  	`json:"category"`
	//можно потом добавить тип тоже
	CreatedAt   time.Time   `json:"created_at"`
}
type TransactionResponse struct {
	ID          uuid.UUID `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Type 		string    `json:"type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int                   `json:"total"`
}