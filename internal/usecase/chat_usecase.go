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

// ChatUseCase handles the AI chat feature.
// It does NOT own any new repo — it borrows existing ones.
type ChatUseCase struct {
    userRepo       repository.UserRepository
    goalRepo       repository.GoalRepository
    statisticsRepo repository.StatisticsRepository
    aiClient       *ai.OpenRouterClient
}

func NewChatUseCase(
    userRepo repository.UserRepository,
    goalRepo repository.GoalRepository,
    statisticsRepo repository.StatisticsRepository,
    aiClient *ai.OpenRouterClient,
) *ChatUseCase {
    return &ChatUseCase{
        userRepo:       userRepo,
        goalRepo:       goalRepo,
        statisticsRepo: statisticsRepo,
        aiClient:       aiClient,
    }
}

// Chat is the main method.
// It fetches live user data → builds context → calls AI → returns reply.
func (uc *ChatUseCase) Chat(ctx context.Context, userID uuid.UUID, history []ai.ChatMessage, userMessage string) (string, error) {
    // 1. Get basic user info (name, currency)
    user, err := uc.userRepo.GetByID(ctx, userID)
    if err != nil {
        return "", fmt.Errorf("chat: get user failed: %w", err)
    }

    // 2. Get current balance
    balance, err := uc.statisticsRepo.GetBalance(ctx, userID)
    if err != nil {
        return "", fmt.Errorf("chat: get balance failed: %w", err)
    }

    // 3. Get last 90 days income & expenses
    now := time.Now()
    start := now.AddDate(0, -3, 0) // 90 days ago
    income, expenses, err := uc.statisticsRepo.GetIncomeExpense(ctx, userID, &start, &now)
    if err != nil {
        return "", fmt.Errorf("chat: get income/expense failed: %w", err)
    }

    // 4. Get top spending categories (last 90 days)
    categories, err := uc.statisticsRepo.GetCategories(ctx, userID, &start, &now)
    if err != nil {
        return "", fmt.Errorf("chat: get categories failed: %w", err)
    }

    // 5. Get all active goals
    goals, err := uc.goalRepo.GetByUserID(ctx, userID)
    if err != nil {
        return "", fmt.Errorf("chat: get goals failed: %w", err)
    }

    // 6. Build the full system prompt (static rules + live data)
    systemPrompt := buildFinancialContext(
        uc.aiClient.GetAdvisorPrompt(),
        user,
        balance,
        income,
        expenses,
        categories,
        goals,
    )

    // 7. Add the current message to the history
    fullHistory := append(history, ai.ChatMessage{
        Role:    "user",
        Content: userMessage,
    })

    // 8. Send to OpenRouter and return the reply
    return uc.aiClient.Chat(ctx, systemPrompt, fullHistory)
}

// buildFinancialContext glues together static advisor rules + live DB data
// into one system prompt string that the AI receives as context.
func buildFinancialContext(
    advisorPrompt string,
    user *domain.User,
    balance *domain.Balance,
    income float64,
    expenses float64,
    categories []*domain.CategoryStats,
    goals []*domain.Goal,
) string {
    var sb strings.Builder

    // --- Static rules from advisor.txt ---
    sb.WriteString(advisorPrompt)
    sb.WriteString("\n\n")

    // --- Live user data from DB ---
    sb.WriteString("=== USER FINANCIAL DATA (live from database) ===\n\n")

    sb.WriteString(fmt.Sprintf("Name: %s\n", user.Name))
    sb.WriteString(fmt.Sprintf("Currency: %s\n\n", user.Currency))

    sb.WriteString(fmt.Sprintf("Current Balance: %.2f %s\n\n", balance.Total, user.Currency))

    sb.WriteString("Last 90 Days Summary:\n")
    sb.WriteString(fmt.Sprintf("  Income:   %.2f %s\n", income, user.Currency))
    sb.WriteString(fmt.Sprintf("  Expenses: %.2f %s\n", expenses, user.Currency))
    sb.WriteString(fmt.Sprintf("  Net Flow: %.2f %s\n\n", income-expenses, user.Currency))

    // Top spending categories (max 5 to avoid token limit)
    expenseCats := filterByType(categories, "expense")
    if len(expenseCats) > 0 {
        sb.WriteString("Top Spending Categories (last 90 days):\n")
        limit := 5
        if len(expenseCats) < limit {
            limit = len(expenseCats)
        }
        for _, cat := range expenseCats[:limit] {
            sb.WriteString(fmt.Sprintf("  - %s: %.2f %s\n", cat.Category, cat.Amount, user.Currency))
        }
        sb.WriteString("\n")
    }

    // Active goals only
    activeGoals := filterActiveGoals(goals)
    if len(activeGoals) > 0 {
        sb.WriteString("Active Savings Goals:\n")
        for _, g := range activeGoals {
            percent := 0.0
            if g.TargetAmount > 0 {
                percent = (g.CurrentAmount / g.TargetAmount) * 100
            }
            remaining := g.TargetAmount - g.CurrentAmount

            line := fmt.Sprintf("  - \"%s\": %.2f / %.2f %s saved (%.1f%% done, %.2f remaining)",
                g.Title, g.CurrentAmount, g.TargetAmount, user.Currency, percent, remaining)

            if g.Deadline != nil {
                daysLeft := int(time.Until(*g.Deadline).Hours() / 24)
                line += fmt.Sprintf(", %d days until deadline", daysLeft)
            } else {
                line += ", no deadline set"
            }
            sb.WriteString(line + "\n")
        }
    } else {
        sb.WriteString("Active Savings Goals: none\n")
    }

    return sb.String()
}

// filterByType returns only categories of a specific type ("expense" or "income")
func filterByType(cats []*domain.CategoryStats, t string) []*domain.CategoryStats {
    var result []*domain.CategoryStats
    for _, c := range cats {
        if c.Type == t {
            result = append(result, c)
        }
    }
    return result
}

// filterActiveGoals returns only goals with status ACTIVE
func filterActiveGoals(goals []*domain.Goal) []*domain.Goal {
    var result []*domain.Goal
    for _, g := range goals {
        if g.Status == domain.GoalStatusActive {
            result = append(result, g)
        }
    }
    return result
}



