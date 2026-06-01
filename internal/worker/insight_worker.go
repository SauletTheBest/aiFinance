package worker

import (
	"context"
	"log"
	"time"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
)

const insightPollInterval = time.Hour

type InsightWorker struct {
	userRepo       repository.UserRepository
	insightUsecase *usecase.InsightUseCase
}

func NewInsightWorker(userRepo repository.UserRepository, insightUsecase *usecase.InsightUseCase) *InsightWorker {
	return &InsightWorker{
		userRepo:       userRepo,
		insightUsecase: insightUsecase,
	}
}

func (w *InsightWorker) Start(ctx context.Context) {
	log.Println("[Worker] AI insight worker started")

	w.refreshAll(ctx)

	ticker := time.NewTicker(insightPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[Worker] AI insight worker stopped")
			return
		case <-ticker.C:
			w.refreshAll(ctx)
		}
	}
}

func (w *InsightWorker) refreshAll(ctx context.Context) {
	userIDs, err := w.userRepo.ListIDs(ctx)
	if err != nil {
		log.Printf("[Worker] failed to list users for insights: %v", err)
		return
	}

	for _, userID := range userIDs {
		if ctx.Err() != nil {
			return
		}

		if _, err := w.insightUsecase.RefreshDueInsights(ctx, userID); err != nil {
			log.Printf("[Worker] failed to refresh insights for user %s: %v", userID, err)
		}
	}
}
