package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/config"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/db"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/handler"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository/postgres"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/server"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/jwt"
)

func main() {

	//поидее тут херня но потом надо это перенести ото че то не чисто да
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	database, err := db.NewPostgres(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate database
	if err := autoMigrate(database); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize services
	jwtSvc := jwt.NewService(cfg.JWTSecret)

	// Initialize repositories
	userRepo := postgres.NewUserRepo(database)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtSvc)

	// Initialize handlers
	authHandler := &handler.AuthHandler{
		AuthUsecase: authUsecase,
	}

	// Setup router
	router := server.SetupRouter(authHandler, jwtSvc)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port https://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&postgres.User{}) //теперь должно быть уникальным
}
