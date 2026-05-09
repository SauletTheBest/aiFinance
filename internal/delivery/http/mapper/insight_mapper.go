package mapper

import (
    "github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
    "github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
)

// ToInsightDTO converts a domain AIInsight into a response DTO
func ToInsightDTO(insight *domain.AIInsight) dto.InsightResponse {
    return dto.InsightResponse{
        ID:          insight.ID.String(),
        Content:     insight.Content,
        InsightType: string(insight.InsightType),
        CreatedAt:   insight.CreatedAt,
    }
}

// ToInsightDTOList converts a slice of domain insights into DTOs
func ToInsightDTOList(insights []*domain.AIInsight) []dto.InsightResponse {
    result := make([]dto.InsightResponse, len(insights))
    for i, ins := range insights {
        result[i] = ToInsightDTO(ins)
    }
    return result
}
