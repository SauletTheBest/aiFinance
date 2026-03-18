package handler


import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/validator"


)

type AuthHandler struct {
	AuthUsecase *usecase.AuthUsecase
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H {
			"error": "invalid request body",
		})
		return
	}

	// Format validation  валидация на слое handler
    if validationErrors := validator.ValidateRegistration(req.Name, req.Email, req.Password); len(validationErrors) > 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "validation failed",
            "details": validationErrors,
        })
		return
	}
	//если пользователь сразу регается то вход идет сразу без повторного логина через login
	token, err := h.AuthUsecase.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {

		// cнова валидация как в usecase
        if err.Error() == "user with this email already exists" {
            c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
            return
        }

		c.JSON(http.StatusInternalServerError, gin.H {
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"user": "user is created",
		"access_token": token,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	token, err := h.AuthUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse {
		AccessToken: token,
	})
}