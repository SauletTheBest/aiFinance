package handler

import (
	"net/http"
	"os"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ParserHandler struct {
	ParserUseCase *usecase.ParserUseCase
}

// UploadStatement handles PDF bank statement upload, parses it, and saves all transactions.
// POST /api/parser/upload
func (h *ParserHandler) UploadStatement(c *gin.Context) {
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

	file, err := c.FormFile("File")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded. Use multipart/form-data with key 'file'"})
		return
	}

	// Validate file extension
	if len(file.Filename) < 4 || file.Filename[len(file.Filename)-4:] != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are accepted"})
		return
	}

	// Save to a temp file, parse it, then clean up
	tempPath := "tmp_kaspi_" + uuid.New().String() + ".pdf"
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}
	defer os.Remove(tempPath)

	transactions, err := h.ParserUseCase.UploadStatement(c.Request.Context(), userID, tempPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Statement parsed and transactions saved successfully",
		"count":   len(transactions),
	})
}
