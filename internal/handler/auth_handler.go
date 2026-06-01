package handler

import (
	"net/http"
	"fmt"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/validator"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthUsecase *usecase.AuthUsecase
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if validationErrors := validator.ValidateRegistration(req.Name, req.Email, req.Password, req.Currency); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": validationErrors,
		})
		return
	}

	err := h.AuthUsecase.Register(c.Request.Context(), req.Name, req.Email, req.Password, req.Currency)
	if err != nil {
		if err.Error() == "user with this email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Return a success message instead of the token!
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Registration successful. Please check your email for the verification code.",
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

// POST /api/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format"})
		return
	}

	// 🆕 Pass the Email and Code directly to the UseCase!
	token, err := h.AuthUsecase.VerifyEmail(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 🆕 Return the token to Flutter!
	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken: token,
	})
}

// POST /api/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}
	err := h.AuthUsecase.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		// Even if it fails, we just say "If email exists, we sent it."
		// This is a security best practice so hackers can't test which emails exist!
		fmt.Printf("Forgot password error: %v\n", err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "If an account with that email exists, a reset code has been sent."})
}
// POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}
	err := h.AuthUsecase.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully! You can now log in."})
}
// POST /api/auth/resend-code
func (h *AuthHandler) ResendVerificationCode(c *gin.Context) {
	var req dto.ResendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	err := h.AuthUsecase.ResendVerificationCode(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "A new verification code has been sent successfully.",
	})
}
