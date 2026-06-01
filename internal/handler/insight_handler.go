package handler

import (
	"net/http"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/mapper"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InsightHandler struct {
	insightUseCase *usecase.InsightUseCase
}

func NewInsightHandler(uc *usecase.InsightUseCase) *InsightHandler {
	return &InsightHandler{insightUseCase: uc}
}

func (h *InsightHandler) GetInsights(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	insights, err := h.insightUseCase.GetLatestInsights(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapper.ToInsightDTOList(insights))
}

func (h *InsightHandler) RefreshInsight(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	typeParam := c.Query("type")
	if typeParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query param 'type' is required",
			"valid": []string{"GOALS", "SPENDING", "GENERAL"},
		})
		return
	}

	insight, err := h.insightUseCase.ForceRefreshInsight(c.Request.Context(), userID, typeParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapper.ToInsightDTO(insight))
}
