package worker

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/ai"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/google/uuid"
)

const (
	// How often the worker checks for PENDING transactions
	pollInterval = 10 * time.Second
	// How many transactions to grab per batch (kept small for free-tier rate limits)
	batchSize = 50
	// Cooldown after a rate-limit (429) or API error before retrying
	errorCooldown = 60 * time.Second
)

// CategorizationWorker is a background goroutine that polls for
// PENDING transactions, sends them to the AI, and saves the categories.
type CategorizationWorker struct {
	transactionRepo repository.TransactionRepository
	aiClient        *ai.OpenRouterClient
}

// NewCategorizationWorker creates a new worker instance.
func NewCategorizationWorker(repo repository.TransactionRepository, aiClient *ai.OpenRouterClient) *CategorizationWorker {
	return &CategorizationWorker{
		transactionRepo: repo,
		aiClient:        aiClient,
	}
}

// Start begins the background polling loop.
// Call this with: go worker.Start(ctx)
//
// Flow per tick:
//  1. Query DB for PENDING transactions (limit = batchSize)
//  2. If none found, sleep and retry
//  3. Set their status to PROCESSING (so no other worker grabs them)
//  4. Send descriptions to AI via OpenRouter
//  5. On success -> update each tx with category + CATEGORIZED status
//  6. On failure -> revert them back to PENDING and wait for cooldown
func (w *CategorizationWorker) Start(ctx context.Context) {
	log.Println("[Worker] Categorization worker started")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[Worker] Categorization worker stopped")
			return
		case <-ticker.C:
			w.processBatch(ctx, ticker)
		}
	}
}

// processBatch handles a single cycle of the worker loop.
func (w *CategorizationWorker) processBatch(ctx context.Context, ticker *time.Ticker) {
	// Step 1: Grab PENDING transactions
	transactions, err := w.transactionRepo.GetByStatus(ctx, domain.StatusPending, batchSize)
	if err != nil {
		log.Printf("[Worker] Error fetching pending transactions: %v", err)
		return
	}

	if len(transactions) == 0 {
		return // nothing to do
	}

	log.Printf("[Worker] Found %d pending transactions to categorize", len(transactions))

	// Step 2: Mark them as PROCESSING
	for _, tx := range transactions {
		tx.Status = domain.StatusProcessing
		if err := w.transactionRepo.Update(ctx, tx); err != nil {
			log.Printf("[Worker] Failed to set PROCESSING for tx %s: %v", tx.ID, err)
		}
	}

	// Step 3: Build the batch map for AI
	items := make(map[uuid.UUID]string, len(transactions))
	for _, tx := range transactions {
		items[tx.ID] = tx.Description
	}

	// Step 4: Call AI
	results, err := w.aiClient.CategorizeTransactions(ctx, items)
	if err != nil {
		log.Printf("[Worker] AI categorization failed: %v", err)

		// Revert back to PENDING so they can be retried later
		w.markAllStatus(ctx, transactions, domain.StatusPending)

		// If rate-limited, pause the worker to avoid hammering
		if isRateLimited(err) {
			log.Printf("[Worker] Rate limited — cooling down for %v", errorCooldown)
			ticker.Reset(errorCooldown)

			// Restore normal interval after one cooldown tick
			go func() {
				time.Sleep(errorCooldown)
				ticker.Reset(pollInterval)
				log.Println("[Worker] Cooldown finished, resuming normal polling")
			}()
		}
		return
	}

	// Step 5: Apply AI results
	resultMap := make(map[uuid.UUID]string, len(results))
	for _, r := range results {
		resultMap[r.TransactionID] = r.Category
	}

	categorized := 0
	for _, tx := range transactions {
		category, ok := resultMap[tx.ID]
		if !ok {
			// AI didn't return a result for this transaction
			tx.Status = domain.StatusFailed
			log.Printf("[Worker] AI did not return category for tx %s", tx.ID)
		} else {
			tx.Category = category
			tx.Status = domain.StatusCategorized
			categorized++
		}

		if err := w.transactionRepo.Update(ctx, tx); err != nil {
			log.Printf("[Worker] Failed to update tx %s: %v", tx.ID, err)
		}
	}

	log.Printf("[Worker] Successfully categorized %d/%d transactions", categorized, len(transactions))
}

// markAllStatus is a helper that sets all given transactions to a specific status.
func (w *CategorizationWorker) markAllStatus(ctx context.Context, transactions []*domain.Transaction, status domain.TransactionStatus) {
	for _, tx := range transactions {
		tx.Status = status
		if err := w.transactionRepo.Update(ctx, tx); err != nil {
			log.Printf("[Worker] Failed to mark tx %s as %s: %v", tx.ID, status, err)
		}
	}
}

// isRateLimited checks if the error message indicates a 429 rate limit.
func isRateLimited(err error) bool {
	return strings.Contains(err.Error(), "429")
}