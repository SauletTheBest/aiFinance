package server

import (
	"github.com/gin-gonic/gin"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/handler"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/middleware"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/jwt"
)

func SetupRouter(authHandler *handler.AuthHandler, userHandler *handler.UserHandler, transactionHandler *handler.TransactionHandler, statisticsHandler *handler.StatisticsHandler, parserHandler *handler.ParserHandler, jwtSvc *jwt.Service) *gin.Engine {
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

		protected.GET("/profile", userHandler.GetProfile)
		protected.POST("/profile", userHandler.UpdateProfile)

		protected.POST("/transactions", transactionHandler.CreateTransaction)
		protected.GET("/transactions/:id", transactionHandler.GetTransaction)
		protected.GET("/transactions", transactionHandler.GetUserTransactions)
		protected.PUT("/transactions/:id", transactionHandler.UpdateTransaction)
		protected.DELETE("/transactions/:id", transactionHandler.DeleteTransaction)

		// Statement upload
		protected.POST("/statements/upload", parserHandler.UploadStatement)

	
		protected.GET("/statistics/balance", statisticsHandler.GetBalance)
		protected.PUT("/statistics/balance", statisticsHandler.UpdateBalance)
		protected.GET("/statistics", statisticsHandler.GetStatistics)
	}

	return router
}
