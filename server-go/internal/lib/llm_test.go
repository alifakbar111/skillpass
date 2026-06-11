package lib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMockLLMClient(t *testing.T) {
	client := NewMockLLMClient()

	var result struct {
		OverallScore int `json:"overallScore"`
	}
	err := client.Chat(context.Background(), "system", "user", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallScore != 75 {
		t.Fatalf("expected 75, got %d", result.OverallScore)
	}
}

func TestNewLLMClientFactory(t *testing.T) {
	t.Run("default returns OpenAI", func(t *testing.T) {
		os.Setenv("LLM_PROVIDER", "")
		defer os.Unsetenv("LLM_PROVIDER")
		client := NewLLMClient()
		if _, ok := client.(*OpenAIClient); !ok {
			t.Fatal("expected OpenAIClient")
		}
	})

	t.Run("anthropic returns Anthropic", func(t *testing.T) {
		os.Setenv("LLM_PROVIDER", "anthropic")
		defer os.Unsetenv("LLM_PROVIDER")
		client := NewLLMClient()
		if _, ok := client.(*AnthropicClient); !ok {
			t.Fatal("expected AnthropicClient")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		os.Setenv("LLM_PROVIDER", "Anthropic")
		defer os.Unsetenv("LLM_PROVIDER")
		client := NewLLMClient()
		if _, ok := client.(*AnthropicClient); !ok {
			t.Fatal("expected AnthropicClient")
		}
	})
}

func TestOpenAIClientChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing auth header")
		}
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": `{"score":42}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		http:    server.Client(),
	}

	var result struct {
		Score int `json:"score"`
	}
	err := client.Chat(context.Background(), "system", "user", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score != 42 {
		t.Fatalf("expected 42, got %d", result.Score)
	}
}

func TestOpenAIClientNoKey(t *testing.T) {
	client := &OpenAIClient{apiKey: ""}
	var result interface{}
	err := client.Chat(context.Background(), "", "", &result)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestAnthropicClientChat(t *testing.T) {
	// Server handles any path (AnthropicClient appends /v1/messages)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("missing x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("missing anthropic-version header")
		}
		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": `{"score":99}`},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		http:    server.Client(),
	}

	var result struct {
		Score int `json:"score"`
	}
	err := client.Chat(context.Background(), "system prompt", "user prompt", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score != 99 {
		t.Fatalf("expected 99, got %d", result.Score)
	}
}

func TestAnthropicClientNoKey(t *testing.T) {
	client := &AnthropicClient{apiKey: ""}
	var result interface{}
	err := client.Chat(context.Background(), "", "", &result)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestAnthropicClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"bad request"}}`))
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		http:    server.Client(),
	}

	var result interface{}
	err := client.Chat(context.Background(), "system", "user", &result)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}
