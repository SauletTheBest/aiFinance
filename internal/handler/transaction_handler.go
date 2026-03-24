package handler

import (
	"net/http"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
)


type TransactionHandler struct {
	TransactionUsecase *usecase.TransactionUsecase
}

func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request												//
	if validationErrors := validator.ValidateTransaction(req.Amount, req.Description, req.Category); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": validationErrors,
		})
		return
	}

	transaction, err := h.TransactionUsecase.CreateTransaction(c.Request.Context(), userID, req.Amount, req.Description, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.TransactionResponse{
		ID:          transaction.ID,
		Amount:      transaction.Amount,
		Description: transaction.Description,
		Category:    transaction.Category,
		Status:      string(transaction.Status),
		CreatedAt:   transaction.CreatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{"transaction": response})
}

func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	transactionID := c.Param("id")
	transactionUUID, err := uuid.Parse(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	transaction, err := h.TransactionUsecase.GetTransaction(c.Request.Context(), transactionUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Ensure user can only access their own transactions
	if transaction.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	response := dto.TransactionResponse{
		ID:          transaction.ID,
		Amount:      transaction.Amount,
		Description: transaction.Description,
		Category:    transaction.Category,
		Status:      string(transaction.Status),
		CreatedAt:   transaction.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"transaction": response})
}

func (h *TransactionHandler) GetUserTransactions(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Pagination parameters maybe need in future 
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	transactions, err := h.TransactionUsecase.GetUserTransactions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit
	if start > len(transactions) {
		start = len(transactions)
	}
	if end > len(transactions) {
		end = len(transactions)
	}

	responseTransactions := make([]dto.TransactionResponse, 0, end-start)
	for _, transaction := range transactions[start:end] {
		responseTransactions = append(responseTransactions, dto.TransactionResponse{
			ID:          transaction.ID,
			Amount:      transaction.Amount,
			Description: transaction.Description,
			Category:    transaction.Category,
			Status:      string(transaction.Status),
			CreatedAt:   transaction.CreatedAt,
		})
	}

	response := dto.TransactionListResponse{
		Transactions: responseTransactions,
		Total:        len(transactions),
	}

	c.JSON(http.StatusOK, response)
}

func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	transactionID := c.Param("id")
	transactionUUID, err := uuid.Parse(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	var req dto.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := validator.ValidateUpdateTransaction(&req.Amount, &req.Description, &req.Category); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": validationErrors,
		})
		return
	}

	// Get existing transaction to ensure it belongs to the user
	existingTransaction, err := h.TransactionUsecase.GetTransaction(c.Request.Context(), transactionUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if existingTransaction.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Update transaction
	err = h.TransactionUsecase.UpdateTransaction(c.Request.Context(), transactionUUID, req.Amount, req.Description, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated successfully"})
}

func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	transactionID := c.Param("id")
	transactionUUID, err := uuid.Parse(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	// Get existing transaction to ensure it belongs to the user
	existingTransaction, err := h.TransactionUsecase.GetTransaction(c.Request.Context(), transactionUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if existingTransaction.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	err = h.TransactionUsecase.DeleteTransaction(c.Request.Context(), transactionUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}


