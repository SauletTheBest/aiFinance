package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/ai"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/google/uuid"
)

// How long before we regenerate each insight type
const (
	goalsRefreshHours    = 12 // Refresh Goals insight every 12 hours
	spendingRefreshHours = 24 // Refresh Spending insight every 24 hours
	generalRefreshHours  = 48 // Refresh General insight every 2 days
)

type InsightUseCase struct {
	insightRepo    repository.InsightRepository
	statisticsRepo repository.StatisticsRepository
	goalRepo       repository.GoalRepository
	aiClient       *ai.OpenRouterClient
}

func NewInsightUseCase(
	insightRepo repository.InsightRepository,
	statisticsRepo repository.StatisticsRepository,
	goalRepo repository.GoalRepository,
	aiClient *ai.OpenRouterClient,
) *InsightUseCase {
	return &InsightUseCase{
		insightRepo:    insightRepo,
		statisticsRepo: statisticsRepo,
		goalRepo:       goalRepo,
		aiClient:       aiClient,
	}
}

// GetOrGenerateInsights returns 3 topic-based insights for the user.
func (uc *InsightUseCase) GetOrGenerateInsights(ctx context.Context, userID uuid.UUID) ([]*domain.AIInsight, error) {
	// Gather ALL user data once (efficient! One DB call, not three)
	now := time.Now()
	start := now.AddDate(0, -1, 0) // Last 30 days of data

	income, expenses, _ := uc.statisticsRepo.GetIncomeExpense(ctx, userID, &start, &now)
	categories, _ := uc.statisticsRepo.GetCategories(ctx, userID, &start, &now)
	balance, _ := uc.statisticsRepo.GetBalance(ctx, userID)
	goals, _ := uc.goalRepo.GetByUserID(ctx, userID)

	// Build one shared data context for all insights
	dataContext := buildInsightContext(income, expenses, balance, categories, goals)

	// Generate all 3 topic insights
	var results []*domain.AIInsight

	topicsAndHours := []struct {
		iType  domain.InsightType
		maxAge float64
	}{
		{domain.InsightTypeGoals, goalsRefreshHours},
		{domain.InsightTypeSpending, spendingRefreshHours},
		{domain.InsightTypeGeneral, generalRefreshHours},
	}

	for _, t := range topicsAndHours {
		insight, err := uc.getOrGenerate(ctx, userID, t.iType, t.maxAge, dataContext)
		if err != nil {
			// Log the error but continue — don't let one failure kill all 3 cards!
			fmt.Printf("warning: failed to generate %s insight: %v\n", t.iType, err)
			continue
		}
		results = append(results, insight)
	}

	return results, nil
}

// getOrGenerate checks DB freshness and generates if needed
func (uc *InsightUseCase) getOrGenerate(ctx context.Context, userID uuid.UUID, iType domain.InsightType, maxAgeHours float64, dataContext string) (*domain.AIInsight, error) {
	// 1. Check if we have a fresh one already
	latest, err := uc.insightRepo.GetLatestByType(ctx, userID, iType)
	if err != nil {
		return nil, err
	}

	if latest != nil {
		hoursSince := time.Since(latest.CreatedAt).Hours()
		if hoursSince < maxAgeHours {
			return latest, nil // Still fresh! Return it.
		}
	}

	// 2. Generate a new one using the topic as context
	prompt := fmt.Sprintf("TOPIC: %s\n\n%s", string(iType), dataContext)
	aiText, err := uc.aiClient.GenerateInsight(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 3. Save and return
	newInsight := &domain.AIInsight{
		ID:          uuid.New(),
		UserID:      userID,
		Content:     aiText,
		InsightType: iType,
		CreatedAt:   time.Now(),
	}
	return newInsight, uc.insightRepo.Create(ctx, newInsight)
}

// buildInsightContext creates a rich data summary for the AI
func buildInsightContext(income, expenses float64, balance *domain.Balance, categories []*domain.CategoryStats, goals []*domain.Goal) string {
	var sb strings.Builder

	if balance != nil {
		sb.WriteString(fmt.Sprintf("Current Balance: %.2f\n", balance.Total))
	}
	sb.WriteString(fmt.Sprintf("Income (last 30 days): %.2f\n", income))
	sb.WriteString(fmt.Sprintf("Expenses (last 30 days): %.2f\n", expenses))
	sb.WriteString(fmt.Sprintf("Net Flow: %.2f\n\n", income-expenses))

	expenseCats := filterByType(categories, "expense")
	if len(expenseCats) > 0 {
		sb.WriteString("Expense Categories:\n")
		for _, c := range expenseCats {
			sb.WriteString(fmt.Sprintf("- %s: %.2f\n", c.Category, c.Amount))
		}
	}

	activeGoals := filterActiveGoals(goals)
	if len(activeGoals) > 0 {
		sb.WriteString("\nActive Goals:\n")
		for _, g := range activeGoals {
			pct := 0.0
			if g.TargetAmount > 0 {
				pct = (g.CurrentAmount / g.TargetAmount) * 100
			}
			sb.WriteString(fmt.Sprintf("- %s: %.0f%% done (saved %.2f of %.2f)\n",
				g.Title, pct, g.CurrentAmount, g.TargetAmount))
		}
	} else {
		sb.WriteString("\nActive Goals: none\n")
	}

	return sb.String()
}
