package handler

import (
	"net/http"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/ai"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
    ChatUseCase *usecase.ChatUseCase
}

type chatRequest struct {
    Message string           `json:"message" binding:"required"`
    History []ai.ChatMessage `json:"history"`
}

// chatResponse is what we send back to Flutter
type chatResponse struct {
    Reply string `json:"reply"`
}

// Chat handles POST /api/ai/chat
func (h *ChatHandler) Chat(c *gin.Context) {
    // 1. Get secure User ID from JWT Token (set by AuthMiddleware)
    userIDStr, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    userID, _ := uuid.Parse(userIDStr.(string))

    // 2. Parse request body
    var req chatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "field 'message' is required"})
        return
    }

    // 3. Call usecase — it handles all the DB fetching + AI call
    reply, err := h.ChatUseCase.Chat(c.Request.Context(), userID, req.History, req.Message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 4. Return the AI reply
    c.JSON(http.StatusOK, chatResponse{Reply: reply})
}


