package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/mapper"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/validator"
)

type StatisticsHandler struct {
	StatisticsUsecase *usecase.StatisticsUsecase
}

func NewStatisticsHandler(uc *usecase.StatisticsUsecase) *StatisticsHandler {
	return &StatisticsHandler{
		StatisticsUsecase: uc,
	}
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, &validator.ValidationError{
			Field:   "authorization",
			Message: "user ID not found in token",
		}
	}

	return uuid.Parse(userIDStr.(string))
}

func writeValidationErrors(c *gin.Context, errs []*validator.ValidationError) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "validation failed",
		"details": errs,
	})
}

func (h *StatisticsHandler) GetBalance(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	balance, err := h.StatisticsUsecase.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := dto.BalanceResponse{
		Total:     balance.Total,
		Currency:  balance.Currency,
		UpdatedAt: balance.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *StatisticsHandler) GetStatistics(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	start, end, validationErrors := validator.ValidatePeriodDates(startStr, endStr)
	if len(validationErrors) > 0 {
		writeValidationErrors(c, validationErrors)
		return
	}

	stats, err := h.StatisticsUsecase.GetStatistics(
		c.Request.Context(),
		userID,
		start,
		end,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := dto.StatisticsResponse{
		Balance: dto.BalanceResponse{
			Total:     stats.Balance.Total,
			Currency:  stats.Balance.Currency,
			UpdatedAt: stats.Balance.UpdatedAt,
		},
		Income:            stats.Income,
		Expenses:          stats.Expenses,
		NetFlow:           stats.NetFlow,
		ExpenseCategories: mapper.ToCategoryDTO(stats.ExpenseCategories),
		IncomeCategories:  mapper.ToCategoryDTO(stats.IncomeCategories),
		PeriodStart:       stats.PeriodStart,
		PeriodEnd:         stats.PeriodEnd,
	}

	c.JSON(http.StatusOK, resp)
}