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
        
        // 💡 Apply the user's opening offset, same as StatisticsUsecase does!
        balance.Total = user.BaseBalance + balance.Total

        now := time.Now()
        last30Start := now.AddDate(0, -1, 0) // 30 days ago
        
        income, expenses, err := uc.statisticsRepo.GetIncomeExpense(ctx, userID, &last30Start, &now)
        if err != nil {
            return "", fmt.Errorf("chat: get income/expense failed: %w", err)
        }

        // 4. Get top spending categories (last 30 days)
        categories, err := uc.statisticsRepo.GetCategories(ctx, userID, &last30Start, &now)
        if err != nil {
            return "", fmt.Errorf("chat: get categories failed: %w", err)
        }

        // 5. Get all active goals
        goals, err := uc.goalRepo.GetByUserID(ctx, userID)
        if err != nil {
            return "", fmt.Errorf("chat: get goals failed: %w", err)
        }

        // ── MVP Metrics ──────────────────────────────────────────
        prev30Start := now.AddDate(0, -2, 0)

        incomePrev, expensesPrev, err := uc.statisticsRepo.GetIncomeExpense(ctx, userID, &prev30Start, &last30Start)
        if err != nil {
            return "", fmt.Errorf("chat: get prev period stats failed: %w", err)
        }

        savingsRate := 0.0
        if income > 0 {
            savingsRate = ((income - expenses) / income) * 100
        }

        incomeTrend := 0.0
        if incomePrev > 0 {
            incomeTrend = ((income - incomePrev) / incomePrev) * 100
        }

        expenseTrend := 0.0
        if expensesPrev > 0 {
            expenseTrend = ((expenses - expensesPrev) / expensesPrev) * 100
        }
        // ── End Metrics ──────────────────────────────────────────

        // 6. Build the full system prompt (static rules + live data)
        systemPrompt := buildFinancialContext(
            uc.aiClient.GetAdvisorPrompt(),
            user,
            balance,
            income,
            expenses,
            categories,
            goals,
            savingsRate,
            incomeTrend,
            expenseTrend,
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
    func buildFinancialContext(
        advisorPrompt string,
        user *domain.User,
        balance *domain.Balance,
        income float64,
        expenses float64,
        categories []*domain.CategoryStats,
        goals []*domain.Goal,
        savingsRate float64,   
        incomeTrend float64,    
        expenseTrend float64,
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

        sb.WriteString("Last 30 Days Summary:\n")
        sb.WriteString(fmt.Sprintf("  Income:   %.2f %s\n", income, user.Currency))
        sb.WriteString(fmt.Sprintf("  Expenses: %.2f %s\n", expenses, user.Currency))
        sb.WriteString(fmt.Sprintf("  Net Flow: %.2f %s\n\n", income-expenses, user.Currency))

        // Top spending categories (max 5 to avoid token limit)
        expenseCats := filterByType(categories, "expense")
        if len(expenseCats) > 0 {
            sb.WriteString("Top Spending Categories (last 30 days):\n")
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

        // --- Computed Metrics ---
        sb.WriteString("\nFinancial Health Metrics:\n")
        sb.WriteString(fmt.Sprintf("  Savings Rate:   %.1f%% (last 30 days)\n", savingsRate))
        sb.WriteString(fmt.Sprintf("  Income Trend:   %+.1f%% (last 30d vs previous 30d)\n", incomeTrend))
        sb.WriteString(fmt.Sprintf("  Expense Trend:  %+.1f%% (last 30d vs previous 30d)\n", expenseTrend))
        sb.WriteString("\nFocus on interpreting these metrics, not repeating raw numbers.\n")

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



