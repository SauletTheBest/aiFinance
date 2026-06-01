package handler

import (
	"net/http"
	"path/filepath"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/ai"
	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	AIClient *ai.OpenRouterClient
}

// ParseVoice handles audio file upload and returns extracted transactions.
// POST /api/ai/voice
func (h *MediaHandler) ParseVoice(c *gin.Context) {
	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No audio file uploaded. Use key 'audio'"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open audio file"})
		return
	}
	defer src.Close()

	// Read all bytes
	audioBytes := make([]byte, file.Size)
	if _, err := src.Read(audioBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio file"})
		return
	}

	// Detect format from file extension
	ext := filepath.Ext(file.Filename)
	format := "wav" // default
	switch ext {
	case ".m4a":
		format = "m4a"
	case ".mp3":
		format = "mp3"
	case ".aac":
		format = "aac"
	case ".ogg":
		format = "ogg"
	}

	// Call AI
	transactions, err := h.AIClient.ParseAudio(c.Request.Context(), audioBytes, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// ParseReceipt handles receipt image upload and returns one extracted transaction.
// POST /api/ai/receipt
func (h *MediaHandler) ParseReceipt(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file uploaded. Use key 'image'"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image file"})
		return
	}
	defer src.Close()

	// Read all bytes
	imageBytes := make([]byte, file.Size)
	if _, err := src.Read(imageBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image file"})
		return
	}

	// Detect MIME type from file extension
	ext := filepath.Ext(file.Filename)
	mimeType := "image/jpeg" // default
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	}

	// Call AI
	transaction, err := h.AIClient.ParseReceiptImage(c.Request.Context(), imageBytes, mimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if transaction == nil {
		c.JSON(http.StatusOK, gin.H{"transaction": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": transaction})
}