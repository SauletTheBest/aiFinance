package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/ai"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/config"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/db"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/handler"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository/postgres"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/server"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/worker"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/jwt"
)

func main() {

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
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services
	jwtSvc := jwt.NewService(cfg.JWTSecret)

	// Initialize repositories
	userRepo := postgres.NewUserRepo(database)
	transactionRepo := postgres.NewTransactionRepo(database)
	statisticsRepo := postgres.NewStatisticsRepo(database)
	goalRepo := postgres.NewGoalRepo(database)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtSvc)
	userUsecase := usecase.NewUserUseCase(userRepo)
	statisticsUsecase := usecase.NewStatisticsUsecase(statisticsRepo, transactionRepo, userRepo)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo, userRepo)
	parserUsecase := usecase.NewParserUseCase(transactionRepo, userRepo)
	goalUsecase := usecase.NewGoalUseCase(goalRepo)

	// Initialize handlers
	authHandler := &handler.AuthHandler{
		AuthUsecase: authUsecase,
	}
	userHandler := &handler.UserHandler{
		UserUsecase: userUsecase,
	}
	transactionHandler := &handler.TransactionHandler{
		TransactionUsecase: transactionUsecase,
	}
	parserHandler := &handler.ParserHandler{
		ParserUseCase: parserUsecase,
	}
	statisticsHandler := &handler.StatisticsHandler{
		StatisticsUsecase: statisticsUsecase,
	}
	goalHandler := &handler.GoalHandler{
		GoalUseCase: goalUsecase,
	}
	// Setup router
	router := server.SetupRouter(authHandler, userHandler, transactionHandler, statisticsHandler, parserHandler, goalHandler, jwtSvc)

	// Initialize AI client and background categorization worker
	aiClient := ai.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
	categorizer := worker.NewCategorizationWorker(transactionRepo, aiClient)

	// Startup recovery: reset any PROCESSING transactions left over from a previous crash
	categorizer.RecoverStuck(context.Background())

	// Create a cancellable context for the worker
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	go categorizer.Start(workerCtx)

	// Graceful shutdown: catch Ctrl+C and SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine so it doesn't block the signal listener
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port https://localhost:%s", port)

	go func() {
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until shutdown signal
	<-quit
	log.Println("[Main] Shutdown signal received — stopping worker...")
	cancelWorker()
	log.Println("[Main] Server stopped.")
}
