package domain

import (
	"time"

	"github.com/google/uuid"
)

type InsightType string

const (
	InsightTypeGoals    InsightType = "GOALS"    // Focuses on savings goals progress
	InsightTypeSpending InsightType = "SPENDING"  // Focuses on spending habits & warnings
	InsightTypeGeneral  InsightType = "GENERAL"   // Overall financial health tip
)

type AIInsight struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Content     string
	InsightType InsightType
	CreatedAt   time.Time
}
