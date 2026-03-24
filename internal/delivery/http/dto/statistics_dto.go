package dto

import "time"

type StatisticsRequest struct {
	PeriodStart *time.Time `json:"period_start,omitempty"`
	PeriodEnd   *time.Time `json:"period_end,omitempty"`
}

type BalanceResponse struct {
	Total     float64 `json:"total"`
	Currency  string  `json:"currency"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryStatsResponse struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
}

type StatisticsResponse struct {
	Balance           BalanceResponse         `json:"balance"`
	Income            float64                 `json:"income"`
	Expenses          float64                 `json:"expenses"`
	NetFlow           float64                 `json:"net_flow"`
	CategoryBreakdown []CategoryStatsResponse `json:"category_breakdown"`
	PeriodStart       time.Time               `json:"period_start"`
	PeriodEnd         time.Time               `json:"period_end"`
}

