package ai

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

//go:embed prompts/categorization.txt
var categorizationPrompt string

//go:embed prompts/advisor.txt
var advisorPrompt string 

//go:embed prompts/insight.txt
var insightPrompt string

//go:embed prompts/voice.txt
var voicePrompt string

//go:embed prompts/receipt.txt
var receiptPrompt string

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


// GenerateInsight asks OpenRouter to create a short financial insight based on provided data context.
func (c *OpenRouterClient) GenerateInsight(ctx context.Context, dataContext string) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: insightPrompt, // Our strict rules
			},
			{
				Role:    "user",
				Content: dataContext,   // The raw numbers we will send from the UseCase
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ai: insight marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("ai: insight request build error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai: insight http error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ai: insight read error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ai: insight status %d: %s", resp.StatusCode, string(respBytes))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("ai: insight parse error: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("ai: insight no choices returned")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ---------- Multimodal structs (for audio and image support) ----------

// ContentPart represents one piece of content: text, image, or audio.
type ContentPart struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	ImageURL   *ImageURL   `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// multimodalRequest is like chatRequest but Content is []ContentPart, not a string.
type multimodalMessage struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

type multimodalRequest struct {
	Model    string               `json:"model"`
	Messages []interface{}        `json:"messages"`
}

// ---------- ParseAudio ----------

// ParseAudio sends an audio file to OpenRouter for transcription and transaction extraction.
// It returns a list of transactions found in the audio.
func (c *OpenRouterClient) ParseAudio(ctx context.Context, audioBytes []byte, format string) ([]map[string]interface{}, error) {
	base64Audio := encodeToBase64(audioBytes)

	reqBody := multimodalRequest{
		Model: c.model,
		Messages: []interface{}{
			ChatMessage{Role: "system", Content: voicePrompt},
			multimodalMessage{
				Role: "user",
				Content: []ContentPart{
					{
						Type:       "input_audio",
						InputAudio: &InputAudio{Data: base64Audio, Format: format},
					},
					{
						Type: "text",
						Text: "Transcribe this audio and extract transactions.",
					},
				},
			},
		},
	}

	return c.sendMultimodalRequest(ctx, reqBody)
}

// ---------- ParseReceiptImage ----------

// ParseReceiptImage sends a receipt photo to OpenRouter for OCR and transaction extraction.
// It returns a single transaction map (or nil if nothing found).
func (c *OpenRouterClient) ParseReceiptImage(ctx context.Context, imageBytes []byte, mimeType string) (map[string]interface{}, error) {
	base64Image := encodeToBase64(imageBytes)
	dataURL := "data:" + mimeType + ";base64," + base64Image

	reqBody := multimodalRequest{
		Model: c.model,
		Messages: []interface{}{
			ChatMessage{Role: "system", Content: receiptPrompt},
			multimodalMessage{
				Role: "user",
				Content: []ContentPart{
					{
						Type:     "image_url",
						ImageURL: &ImageURL{URL: dataURL},
					},
					{
						Type: "text",
						Text: "Recognize this receipt and create one transaction for the total amount.",
					},
				},
			},
		},
	}

	results, err := c.sendMultimodalRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// ---------- Shared helpers ----------

// encodeToBase64 converts raw bytes into a base64 string.
func encodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// sendMultimodalRequest sends the request to OpenRouter and parses the JSON array response.
func (c *OpenRouterClient) sendMultimodalRequest(ctx context.Context, reqBody multimodalRequest) ([]map[string]interface{}, error) {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ai: failed to marshal multimodal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ai: failed to build multimodal request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai: multimodal http error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ai: multimodal read error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai: multimodal status %d: %s", resp.StatusCode, string(respBytes))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("ai: multimodal parse error: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("ai: multimodal no choices returned")
	}

	rawContent := cleanJSONResponse(chatResp.Choices[0].Message.Content)
	var voiceResponse struct {
		Transcription string                   `json:"transcription"`
		Transactions  []map[string]interface{} `json:"transactions"`
	}
	if err := json.Unmarshal([]byte(rawContent), &voiceResponse); err == nil && voiceResponse.Transcription != "" {
		
		for _, tx := range voiceResponse.Transactions {
			tx["_transcription"] = voiceResponse.Transcription
		}
		return voiceResponse.Transactions, nil
	}

	// Try to parse as object first (receipt returns single object)
	var single map[string]interface{}
	if err := json.Unmarshal([]byte(rawContent), &single); err == nil {
		return []map[string]interface{}{single}, nil
	}

	// Try to parse as array (voice returns array of transactions)
	var many []map[string]interface{}
	if err := json.Unmarshal([]byte(rawContent), &many); err == nil {
		return many, nil
	}

	return nil, fmt.Errorf("ai: could not parse response as JSON: %s", rawContent)
}