package dto

import "time"

type InsightResponse struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	InsightType string    `json:"type"` // "DAILY", "WEEKLY", "MONTHLY"
	CreatedAt   time.Time `json:"created_at"`
}
