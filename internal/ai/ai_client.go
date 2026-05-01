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
	"strings"
	"github.com/google/uuid"
)

//go:embed prompts/categorization.txt
var categorizationPrompt string

//go:embed prompts/advisor.txt
var advisorPrompt string 

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

// helper function to clean ai response
func cleanJSONResponse(raw string) string {
	// Remove the ```json at the start
	raw = strings.TrimPrefix(raw, "```json")
	// Remove the ``` at the start (in case the AI didn't include "json")
	raw = strings.TrimPrefix(raw, "```")
	// Remove the ``` at the end
	raw = strings.TrimSuffix(raw, "```")
	// Remove any extra whitespace
	return strings.TrimSpace(raw)
}

// ---------- OpenRouter request / response structs ----------

// ChatMessage represents a single message in the conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string         `json:"model"`
	Messages []ChatMessage `json:"messages"`
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
		Messages: []ChatMessage{
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

	cleanContent := cleanJSONResponse(rawContent)
	
	var categoryMap map[string]string

	if err := json.Unmarshal([]byte(cleanContent), &categoryMap); err != nil {
		return nil, fmt.Errorf("ai: failed to parse AI output as JSON: %w\nRaw: %s", err, cleanContent)
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


// Chat sends a user message to OpenRouter with a pre-built financial context and history.
func (c *OpenRouterClient) Chat(ctx context.Context, systemPrompt string, history []ChatMessage) (string, error) {
	// Build the message list: System Prompt + History
	messages := make([]ChatMessage, 0, len(history)+1)
	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})
	messages = append(messages, history...)

	reqBody := chatRequest{
		Model:    c.model,
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ai: chat marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("ai: chat request build error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai: chat http error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ai: chat read error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ai: chat status %d: %s", resp.StatusCode, string(respBytes))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("ai: chat parse error: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("ai: chat no choices returned")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// GetAdvisorPrompt returns the static advisor rules loaded from advisor.txt
func (c *OpenRouterClient) GetAdvisorPrompt() string {
    return advisorPrompt
}
