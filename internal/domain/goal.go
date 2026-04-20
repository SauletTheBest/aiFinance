package domain

import (
	"time"

	"github.com/google/uuid"
)

type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "ACTIVE"
	GoalStatusCompleted GoalStatus = "COMPLETED"
	GoalStatusCancelled GoalStatus = "CANCELLED"
)

type Goal struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Title         string
	TargetAmount  float64
	CurrentAmount float64
	Deadline      *time.Time
	Status        GoalStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
