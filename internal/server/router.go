package server

import (
	"github.com/gin-gonic/gin"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/handler"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/middleware"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/jwt"
)

func SetupRouter(authHandler *handler.AuthHandler, userHandler *handler.UserHandler, jwtSvc *jwt.Service) *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Auth routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(jwtSvc))
	{
		// Add protected routes here
		// Example: protected.GET("/profile", profileHandler.GetProfile)

		protected.GET("/profile", userHandler.GetProfile)
		protected.POST("/profile", userHandler.UpdateProfile)
	}

	return router
}
