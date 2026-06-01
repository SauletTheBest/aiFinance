package mapper

import (
	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
)

func ToInsightDTO(insight *domain.AIInsight) dto.InsightResponse {
	return dto.InsightResponse{
		ID:          insight.ID.String(),
		Content:     insight.Content,
		InsightType: string(insight.InsightType),
		CreatedAt:   insight.CreatedAt,
	}
}

func ToInsightDTOList(insights []*domain.AIInsight) []dto.InsightResponse {
	result := make([]dto.InsightResponse, len(insights))
	for i, insight := range insights {
		result[i] = ToInsightDTO(insight)
	}
	return result
}
