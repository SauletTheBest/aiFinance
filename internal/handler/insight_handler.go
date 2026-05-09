package handler

import (
	"net/http"

	//"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
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

// GET /api/insights
func (h *InsightHandler) GetInsights(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	insights, err := h.insightUseCase.GetOrGenerateInsights(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// mapper converts domain → dto. Handler knows nothing about domain! ✅
	c.JSON(http.StatusOK, mapper.ToInsightDTOList(insights))
}

// POST /api/insights/refresh?type=GOALS
func (h *InsightHandler) RefreshInsight(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	// Just a plain string — no domain import needed! ✅
	typeParam := c.Query("type")
	if typeParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query param 'type' is required",
			"valid": []string{"GOALS", "SPENDING", "GENERAL"},
		})
		return
	}

	// Pass string to UseCase — UseCase converts to domain.InsightType internally
	insight, err := h.insightUseCase.ForceRefreshInsight(c.Request.Context(), userID, typeParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// mapper converts domain → dto ✅
	c.JSON(http.StatusOK, mapper.ToInsightDTO(insight))
}
