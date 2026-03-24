package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
)

type StatisticsHandler struct {
	StatisticsUsecase *usecase.StatisticsUsecase
}

// helper — убирает дублирование получения user_id(вывел в функцию но можно потом в уровня пакета)
func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, http.ErrNoCookie
	}

	return uuid.Parse(userIDStr.(string))
}

func (h *StatisticsHandler) GetBalance(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	balance, err := h.StatisticsUsecase.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.BalanceResponse{
		Total:     balance.Total,
		Currency:  balance.Currency,
		UpdatedAt: balance.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"balance": response})
}

func (h *StatisticsHandler) GetStatistics(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.StatisticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.StatisticsUsecase.GetStatistics(
		c.Request.Context(),
		userID,
		req.PeriodStart,
		req.PeriodEnd,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.StatisticsResponse{
		Balance: dto.BalanceResponse{
			Total:     stats.Balance.Total,
			Currency:  stats.Balance.Currency,
			UpdatedAt: stats.Balance.UpdatedAt,
		},
		Income:      stats.Income,
		Expenses:    stats.Expenses,
		NetFlow:     stats.NetFlow,
		PeriodStart: stats.PeriodStart,
		PeriodEnd:   stats.PeriodEnd,
	}

	response.CategoryBreakdown = make([]dto.CategoryStatsResponse, len(stats.CategoryBreakdown))
	for i, category := range stats.CategoryBreakdown {
		response.CategoryBreakdown[i] = dto.CategoryStatsResponse{
			Category: category.Category,
			Amount:   category.Amount,
			Count:    category.Count,
		}
	}

	c.JSON(http.StatusOK, gin.H{"statistics": response})
}

func (h *StatisticsHandler) GetCategoryBreakdown(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.StatisticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	categoryStats, err := h.StatisticsUsecase.GetCategoryBreakdown(
		c.Request.Context(),
		userID,
		req.PeriodStart,
		req.PeriodEnd,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.CategoryStatsResponse, len(categoryStats))
	for i, category := range categoryStats {
		response[i] = dto.CategoryStatsResponse{
			Category: category.Category,
			Amount:   category.Amount,
			Count:    category.Count,
		}
	}

	c.JSON(http.StatusOK, gin.H{"category_breakdown": response})
}