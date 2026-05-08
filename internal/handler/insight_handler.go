package handler

import (
	"net/http"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InsightHandler struct {
	insightUseCase *usecase.InsightUseCase
}

func NewInsightHandler(uc *usecase.InsightUseCase) *InsightHandler {
	return &InsightHandler{
		insightUseCase: uc,
	}
}

// GetInsights handles GET /api/insights
func (h *InsightHandler) GetInsights(c *gin.Context) {
	// 1. Check if user is logged in (from JWT middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	// 2. Ask UseCase to get or generate the insights
	insights, err := h.insightUseCase.GetOrGenerateInsights(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Map Domain models to DTOs for Flutter
	var response []dto.InsightResponse
	for _, ins := range insights {
		response = append(response, dto.InsightResponse{
			ID:          ins.ID.String(),
			Content:     ins.Content,
			InsightType: string(ins.InsightType),
			CreatedAt:   ins.CreatedAt,
		})
	}

	// 4. Send JSON back!
	c.JSON(http.StatusOK, response)
}
