package mapper

import (
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
)

func ToCategoryDTO(categories []*domain.CategoryStats) []dto.CategoryStatsResponse {
	result := make([]dto.CategoryStatsResponse, len(categories))

	for i, c := range categories {
		result[i] = dto.CategoryStatsResponse{
			Category: c.Category,
			Amount:   c.Amount,
			Count:    c.Count,
		}
	}

	return result
}