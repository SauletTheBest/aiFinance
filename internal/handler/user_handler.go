package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
)

type UserHandler struct {
	UserUsecase *usecase.UserUseCase
}

func (u *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
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

	user, err := u.UserUsecase.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ProfileResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Currency:  user.Currency,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"profile": response})
}

func (u *UserHandler) UpdateProfile(c *gin.Context) {
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

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = u.UserUsecase.UpdateProfile(c.Request.Context(), userID, req.Name, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}
