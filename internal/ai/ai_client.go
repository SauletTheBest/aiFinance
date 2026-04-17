package ai

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

//go:embed prompts/categorization.txt
var categorizationPrompt string

// CategoryResult holds the AI's categorization decision for a single transaction.
type CategoryResult struct {
	TransactionID uuid.UUID
	Category      string
}

// OpenRouterClient communicates with the OpenRouter API to categorize transactions.
type OpenRouterClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenRouterClient creates a new AI client.
func NewOpenRouterClient(apiKey string, model string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ---------- OpenRouter request / response structs ----------

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ---------- Core categorization logic ----------

// CategorizeTransactions takes a batch of (id, description) pairs
// and returns a map of transaction_id -> category.
//
// The AI prompt instructs the model to act as a financial analyst:
//   - "Magnum"       -> "Groceries"
//   - "Yandex Go"    -> "Transport"
//   - "Fashion LLP"  -> "Clothing"
//
// The model must respond with pure JSON: { "<uuid>": "<category>", ... }
func (c *OpenRouterClient) CategorizeTransactions(ctx context.Context, items map[uuid.UUID]string) ([]CategoryResult, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// Build the prompt
	prompt := buildPrompt(items)

	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: categorizationPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ai: failed to marshal request: %w", err)
	}

	// Build HTTP request to OpenRouter
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ai: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Execute
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ai: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai: OpenRouter returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	// Parse the OpenRouter chat response
	var chatResp chatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("ai: failed to parse response JSON: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("ai: no choices in response")
	}

	// The AI content should be a pure JSON object: { "uuid": "category", ... }
	rawContent := chatResp.Choices[0].Message.Content
	var categoryMap map[string]string
	if err := json.Unmarshal([]byte(rawContent), &categoryMap); err != nil {
		return nil, fmt.Errorf("ai: failed to parse AI output as JSON: %w\nRaw: %s", err, rawContent)
	}

	// Map back to CategoryResult
	results := make([]CategoryResult, 0, len(categoryMap))
	for idStr, category := range categoryMap {
		txID, err := uuid.Parse(idStr)
		if err != nil {
			continue // skip malformed IDs from AI
		}
		results = append(results, CategoryResult{
			TransactionID: txID,
			Category:      category,
		})
	}

	return results, nil
}

// buildPrompt creates a JSON string of { "tx_id": "description", ... }
func buildPrompt(items map[uuid.UUID]string) string {
	data, _ := json.MarshalIndent(items, "", "  ")
	return string(data)
}