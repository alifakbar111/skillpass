package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// LLMClient sends prompts to an LLM and returns structured JSON responses.
type LLMClient interface {
	// Chat sends a prompt and expects a JSON response matching the provided schema.
	// The systemPrompt sets the LLM's behavior; userPrompt is the specific input.
	// The response is unmarshalled into the resultPtr (must be a pointer to a struct).
	Chat(ctx context.Context, systemPrompt, userPrompt string, resultPtr interface{}) error
}

// Compile-time interface checks.
var _ LLMClient = (*OpenAIClient)(nil)
var _ LLMClient = (*AnthropicClient)(nil)
var _ LLMClient = (*MockLLMClient)(nil)

// NewLLMClient creates an LLMClient based on the LLM_PROVIDER env var.
// Supported values: "anthropic", "openai" (default).
func NewLLMClient() LLMClient {
	provider := os.Getenv("LLM_PROVIDER")
	switch strings.ToLower(provider) {
	case "anthropic":
		return NewAnthropicClient()
	default:
		return NewOpenAIClient()
	}
}

// OpenAIClient implements LLMClient using the OpenAI-compatible chat completions API.
type OpenAIClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// NewOpenAIClient creates an LLM client from environment variables:
// LLM_API_KEY (required), LLM_MODEL (default: gpt-4o-mini), LLM_BASE_URL (default: https://api.openai.com/v1).
func NewOpenAIClient() *OpenAIClient {
	apiKey := os.Getenv("LLM_API_KEY")
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &OpenAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature"`
	MaxTokens      int             `json:"max_tokens"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAIClient) Chat(ctx context.Context, systemPrompt, userPrompt string, resultPtr interface{}) error {
	if c.apiKey == "" {
		return fmt.Errorf("LLM_API_KEY not configured — cannot call LLM")
	}

	body := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature:    0.3,
		MaxTokens:      4096,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyPreview, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("llm API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyPreview)))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}

	if chatResp.Error != nil {
		return fmt.Errorf("llm API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return fmt.Errorf("llm returned no choices (body: %s)", string(respBody))
	}

	content := chatResp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), resultPtr); err != nil {
		return fmt.Errorf("parse llm json response: %w (content: %s)", err, content)
	}

	return nil
}

// AnthropicClient implements LLMClient using the Anthropic Messages API.
type AnthropicClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// NewAnthropicClient creates an LLM client for the Anthropic API from environment variables:
// LLM_API_KEY (required), LLM_MODEL (default: claude-sonnet-4-6), LLM_BASE_URL (default: https://api.anthropic.com).
func NewAnthropicClient() *AnthropicClient {
	apiKey := os.Getenv("LLM_API_KEY")
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &AnthropicClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *AnthropicClient) Chat(ctx context.Context, systemPrompt, userPrompt string, resultPtr interface{}) error {
	if c.apiKey == "" {
		return fmt.Errorf("LLM_API_KEY not configured — cannot call Anthropic API")
	}

	body := anthropicRequest{
		Model:     c.model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt + "\n\nRespond with valid JSON only."},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyPreview, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("anthropic API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyPreview)))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}

	if anthropicResp.Error != nil {
		return fmt.Errorf("anthropic API error: %s", anthropicResp.Error.Message)
	}

	// Extract text content from response blocks
	var textContent string
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			textContent = block.Text
			break
		}
	}
	if textContent == "" {
		return fmt.Errorf("anthropic returned no text content (body: %s)", string(respBody))
	}

	if err := json.Unmarshal([]byte(textContent), resultPtr); err != nil {
		return fmt.Errorf("parse anthropic json response: %w (content: %s)", err, textContent)
	}

	return nil
}

// MockLLMClient implements LLMClient returning predefined responses (for dev/testing).
type MockLLMClient struct {
	ResponseFunc func(systemPrompt, userPrompt string) interface{}
}

func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		ResponseFunc: func(systemPrompt, userPrompt string) interface{} {
			return map[string]interface{}{
				"overallScore": 75,
				"strengths":    []map[string]interface{}{{"skill": "Go", "score": 90, "note": "Strong backend skills"}},
				"weaknesses":   []map[string]interface{}{{"skill": "React", "score": 40, "note": "Limited frontend experience"}},
				"suggestions":  []map[string]interface{}{{"area": "Frontend", "tip": "Build a React project"}},
				"skillScores":  []map[string]interface{}{{"skill": "Go", "category": "backend", "score": 90}},
			}
		},
	}
}

func (m *MockLLMClient) Chat(ctx context.Context, systemPrompt, userPrompt string, resultPtr interface{}) error {
	result := m.ResponseFunc(systemPrompt, userPrompt)
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("mock marshal: %w", err)
	}
	return json.Unmarshal(data, resultPtr)
}
