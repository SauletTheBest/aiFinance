package dto

import "time"

type StatisticsRequest struct {
    PeriodStart string `form:"period_start" json:"period_start,omitempty"`
    PeriodEnd   string `form:"period_end" json:"period_end,omitempty"`
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
	ExpenseCategories []CategoryStatsResponse `json:"expense_categories"`
	IncomeCategories  []CategoryStatsResponse `json:"income_categories"`
	PeriodStart       *time.Time               `json:"period_start,omitempty"`
	PeriodEnd         *time.Time               `json:"period_end,omitempty"`
}

