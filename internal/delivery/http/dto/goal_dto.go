package dto

import (
	"time"
	"github.com/google/uuid"
)

// What the frontend sends us to create a goal
type CreateGoalRequest struct {
	Title        string     `json:"title" binding:"required"`
	Description  *string    `json:"description"`
	TargetAmount float64    `json:"target_amount" binding:"required,gt=0"`
	Deadline     *time.Time `json:"deadline"`
}

// What the frontend sends us to add money
type AddProgressRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// What we send back to the frontend
type GoalResponse struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	TargetAmount  float64    `json:"target_amount"`
	CurrentAmount float64    `json:"current_amount"`
	Deadline      *time.Time `json:"deadline"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}
