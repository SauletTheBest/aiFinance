package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/ai"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/config"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/db"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/email"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/handler"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository/postgres"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/server"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/worker"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type App struct {
	router      *gin.Engine
	categorizer *worker.CategorizationWorker
	port        string
}

func NewApp(cfg *config.Config) *App {

	//data sourse name
	dsn := fmt.Sprintf("host=%s port= %d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	database, err := db.NewPostgres(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	aiClient := ai.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)

	jwtSvc := jwt.NewService(cfg.JWTSecret)

	userRepo := postgres.NewUserRepo(database)
	transactionRepo := postgres.NewTransactionRepo(database)
	statisticsRepo := postgres.NewStatisticsRepo(database)
	goalRepo := postgres.NewGoalRepo(database)
	insightRepo := postgres.NewInsightRepo(database)
	verificationRepo := postgres.NewVerificationRepo(database)

	emailSvc := email.NewEmailService(cfg.GmailClientID, cfg.GmailClientSecret, cfg.GmailRefreshToken, cfg.GmailSender)

	authUsecase := usecase.NewAuthUsecase(userRepo, jwtSvc, verificationRepo, emailSvc)
	userUsecase := usecase.NewUserUseCase(userRepo)
	statisticsUsecase := usecase.NewStatisticsUsecase(statisticsRepo, transactionRepo, userRepo)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo, userRepo)
	parserUsecase := usecase.NewParserUseCase(transactionRepo, userRepo)
	goalUsecase := usecase.NewGoalUseCase(goalRepo)
	insightUseCase := usecase.NewInsightUseCase(insightRepo, statisticsRepo, goalRepo, aiClient)

	chatUsecase := usecase.NewChatUseCase(userRepo, goalRepo, statisticsRepo, aiClient)

	authHandler := &handler.AuthHandler{AuthUsecase: authUsecase}
	userHandler := &handler.UserHandler{UserUsecase: userUsecase}
	transactionHandler := &handler.TransactionHandler{TransactionUsecase: transactionUsecase}
	parserHandler := &handler.ParserHandler{ParserUseCase: parserUsecase}
	statisticsHandler := &handler.StatisticsHandler{StatisticsUsecase: statisticsUsecase}
	goalHandler := &handler.GoalHandler{GoalUseCase: goalUsecase}
	insightHandler := handler.NewInsightHandler(insightUseCase)

	chatHandler := &handler.ChatHandler{ChatUseCase: chatUsecase}

	router := server.SetupRouter(authHandler, userHandler, transactionHandler, statisticsHandler, parserHandler, goalHandler, chatHandler, insightHandler, jwtSvc)

	categorizer := worker.NewCategorizationWorker(transactionRepo, aiClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &App{
		router:      router,
		categorizer: categorizer,
		port:        port,
	}
}

func (a *App) Run() {

	// Startup recovery: reset any PROCESSING transactions left over from a previous crash
	a.categorizer.RecoverStuck(context.Background())

	// Create a cancellable context for the worker
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	go a.categorizer.Start(workerCtx)

	// Graceful shutdown: catch Ctrl+C and SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	log.Printf("Server starting on port https://localhost:%s", a.port)
	go func() {
		if err := a.router.Run(":" + a.port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until shutdown signal
	<-quit
	log.Println("[App] Shutdown signal received — stopping worker...")
	cancelWorker()
	log.Println("[App] Server stopped.")
}
