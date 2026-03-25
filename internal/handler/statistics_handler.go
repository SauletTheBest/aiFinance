package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"time"
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
	// 1. Получаем user_id из JWT
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 2. Получаем параметры запроса
	periodStart, periodEnd, err := h.parsePeriodParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// 3. Вызываем бизнес-логику
	stats, err := h.StatisticsUsecase.GetStatistics(
		c.Request.Context(),
		userID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	expenseCategories := make([]dto.CategoryStatsResponse, len(stats.ExpenseCategories))
    for i, category := range stats.ExpenseCategories {
        expenseCategories[i] = dto.CategoryStatsResponse{
            Category: category.Category,  
            Amount:   category.Amount,    
            Count:    category.Count,     
        }
    }
    
    incomeCategories := make([]dto.CategoryStatsResponse, len(stats.IncomeCategories))
    for i, category := range stats.IncomeCategories {
        incomeCategories[i] = dto.CategoryStatsResponse{
            Category: category.Category, 
            Amount:   category.Amount,  
            Count:    category.Count,  
        }
    }

	// 4. Возвращаем результат
	c.JSON(http.StatusOK, gin.H{
        "balance": gin.H{
            "total":      stats.Balance.Total,
            "currency":   stats.Balance.Currency,
            "updated_at": stats.Balance.UpdatedAt,
        },
        "income":   stats.Income,
        "expenses": stats.Expenses,
        "net_flow": stats.NetFlow,
        
        "expense_categories": expenseCategories, 
        "income_categories":  incomeCategories,  
        
        "period_start": stats.PeriodStart,
        "period_end":   stats.PeriodEnd,
    })
}


func (h *StatisticsHandler) parsePeriodParams(c *gin.Context) (*time.Time, *time.Time, error) {
	var periodStart, periodEnd *time.Time

	if start := c.Query("start"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			periodStart = &t
		} else {
			return nil, nil, err
		}
	}

	if end := c.Query("end"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			periodEnd = &t
		} else {
			return nil, nil, err
		}
	}

	return periodStart, periodEnd, nil
}