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

const (
	goalsRefreshHours    = 12
	spendingRefreshHours = 24
	generalRefreshHours  = 48
)

type insightTopic struct {
	iType       domain.InsightType
	maxAgeHours float64
}

type InsightUseCase struct {
	insightRepo    repository.InsightRepository
	statisticsRepo repository.StatisticsRepository
	goalRepo       repository.GoalRepository
	userRepo       repository.UserRepository
	aiClient       *ai.OpenRouterClient
}

func NewInsightUseCase(
	insightRepo repository.InsightRepository,
	statisticsRepo repository.StatisticsRepository,
	goalRepo repository.GoalRepository,
	userRepo repository.UserRepository,
	aiClient *ai.OpenRouterClient,
) *InsightUseCase {
	return &InsightUseCase{
		insightRepo:    insightRepo,
		statisticsRepo: statisticsRepo,
		goalRepo:       goalRepo,
		userRepo:       userRepo,
		aiClient:       aiClient,
	}
}

func (uc *InsightUseCase) GetOrGenerateInsights(ctx context.Context, userID uuid.UUID) ([]*domain.AIInsight, error) {
	return uc.GetLatestInsights(ctx, userID)
}

// GetLatestInsights only reads saved insights. It never calls AI from a page request.
func (uc *InsightUseCase) GetLatestInsights(ctx context.Context, userID uuid.UUID) ([]*domain.AIInsight, error) {
	results := make([]*domain.AIInsight, 0, len(insightTopics()))

	for _, topic := range insightTopics() {
		insight, err := uc.insightRepo.GetLatestByType(ctx, userID, topic.iType)
		if err != nil {
			return nil, err
		}
		if insight != nil {
			results = append(results, insight)
		}
	}

	return results, nil
}

// RefreshDueInsights is called by the background worker. It saves new rows only
// when an insight is missing or older than its configured refresh period.
func (uc *InsightUseCase) RefreshDueInsights(ctx context.Context, userID uuid.UUID) ([]*domain.AIInsight, error) {
	results := make([]*domain.AIInsight, 0, len(insightTopics()))
	dataContext, err := uc.buildUserInsightContext(ctx, userID)
	if err != nil {
		return nil, err
	}

	var firstErr error
	for _, topic := range insightTopics() {
		latest, err := uc.insightRepo.GetLatestByType(ctx, userID, topic.iType)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		if latest != nil && time.Since(latest.CreatedAt).Hours() < topic.maxAgeHours {
			results = append(results, latest)
			continue
		}

		insight, err := uc.generateAndSaveInsight(ctx, userID, topic.iType, dataContext)
		if err != nil {
			if latest != nil {
				results = append(results, latest)
			}
			if firstErr == nil {
				firstErr = fmt.Errorf("refresh %s insight: %w", topic.iType, err)
			}
			continue
		}
		results = append(results, insight)
	}

	return results, firstErr
}

// ForceRefreshInsight generates a brand new insight regardless of cache.
func (uc *InsightUseCase) ForceRefreshInsight(ctx context.Context, userID uuid.UUID, typeStr string) (*domain.AIInsight, error) {
	iType, err := parseInsightType(typeStr)
	if err != nil {
		return nil, err
	}

	dataContext, err := uc.buildUserInsightContext(ctx, userID)
	if err != nil {
		return nil, err
	}

	insight, err := uc.generateAndSaveInsight(ctx, userID, iType, dataContext)
	if err != nil {
		return nil, fmt.Errorf("refresh %s insight: %w", iType, err)
	}

	return insight, nil
}

func (uc *InsightUseCase) generateAndSaveInsight(ctx context.Context, userID uuid.UUID, iType domain.InsightType, dataContext string) (*domain.AIInsight, error) {
	prompt := fmt.Sprintf("TOPIC: %s\n\n%s", string(iType), dataContext)
	aiText, err := uc.aiClient.GenerateInsight(ctx, prompt)
	if err != nil {
		return nil, err
	}

	newInsight := &domain.AIInsight{
		ID:          uuid.New(),
		UserID:      userID,
		Content:     aiText,
		InsightType: iType,
		CreatedAt:   time.Now(),
	}
	if err := uc.insightRepo.Create(ctx, newInsight); err != nil {
		return nil, err
	}

	return newInsight, nil
}

func (uc *InsightUseCase) buildUserInsightContext(ctx context.Context, userID uuid.UUID) (string, error) {
	now := time.Now()
	start := now.AddDate(0, -1, 0)

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	income, expenses, err := uc.statisticsRepo.GetIncomeExpense(ctx, userID, &start, &now)
	if err != nil {
		return "", err
	}
	categories, err := uc.statisticsRepo.GetCategories(ctx, userID, &start, &now)
	if err != nil {
		return "", err
	}
	balance, err := uc.statisticsRepo.GetBalance(ctx, userID)
	if err != nil {
		return "", err
	}
	balance.Total = user.BaseBalance + balance.Total
	balance.Currency = user.Currency
	goals, err := uc.goalRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	return buildInsightContext(income, expenses, balance, categories, goals), nil
}

func insightTopics() []insightTopic {
	return []insightTopic{
		{domain.InsightTypeGoals, goalsRefreshHours},
		{domain.InsightTypeSpending, spendingRefreshHours},
		{domain.InsightTypeGeneral, generalRefreshHours},
	}
}

func parseInsightType(typeStr string) (domain.InsightType, error) {
	switch domain.InsightType(typeStr) {
	case domain.InsightTypeGoals:
		return domain.InsightTypeGoals, nil
	case domain.InsightTypeSpending:
		return domain.InsightTypeSpending, nil
	case domain.InsightTypeGeneral:
		return domain.InsightTypeGeneral, nil
	default:
		return "", fmt.Errorf("invalid insight type %q", typeStr)
	}
}

func buildInsightContext(income, expenses float64, balance *domain.Balance, categories []*domain.CategoryStats, goals []*domain.Goal) string {
	var sb strings.Builder

	if balance != nil {
		sb.WriteString(fmt.Sprintf("Current Balance: %.2f %s\n", balance.Total, balance.Currency))
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
