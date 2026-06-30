package lib

import (
	"bytes"
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

func TestExtractOuterJSON(t *testing.T) {
	t.Run("strips data DONE trailing junk", func(t *testing.T) {
		body := []byte(`{"model":"claude-haiku-4-5-20251001","content":[{"type":"text","text":"hello"}]}data: [DONE]`)
		got := extractOuterJSON(body)
		want := `{"model":"claude-haiku-4-5-20251001","content":[{"type":"text","text":"hello"}]}`
		if string(got) != want {
			t.Fatalf("got %q, want %q", string(got), want)
		}
	})

	t.Run("no trailing junk", func(t *testing.T) {
		body := []byte(`{"key":"value"}`)
		got := extractOuterJSON(body)
		if string(got) != string(body) {
			t.Fatalf("got %q, want %q", string(got), string(body))
		}
	})

	t.Run("whitespace around JSON", func(t *testing.T) {
		body := []byte("  \n  {\"key\":\"value\"}  \n  ")
		got := extractOuterJSON(body)
		want := `{"key":"value"}`
		if string(got) != want {
			t.Fatalf("got %q, want %q", string(got), want)
		}
	})
}

func TestStripMarkdownFences(t *testing.T) {
	t.Run("json code block", func(t *testing.T) {
		input := "```json\n{\"score\":42}\n```"
		got := stripMarkdownFences(input)
		want := `{"score":42}`
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("no fences", func(t *testing.T) {
		input := `{"score":42}`
		got := stripMarkdownFences(input)
		if got != input {
			t.Fatalf("got %q, want %q", got, input)
		}
	})

	t.Run("plain code block", func(t *testing.T) {
		input := "```\nplain text\n```"
		got := stripMarkdownFences(input)
		want := "plain text"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func proxyResponse(t *testing.T, textContent string, suffix string) []byte {
	t.Helper()
	body := map[string]interface{}{
		"model":       "claude-haiku-4-5-20251001",
		"id":          "msg_test",
		"type":        "message",
		"role":        "assistant",
		"stop_reason": "end_turn",
		"usage":       map[string]interface{}{},
		"content": []map[string]interface{}{
			{"type": "text", "text": textContent},
		},
	}
	j, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	buf.Write(j)
	buf.WriteString(suffix)
	return buf.Bytes()
}

func TestAnthropicClientProxyResponse(t *testing.T) {
	// Simulate the 9Router proxy returning Anthropic format with:
	// - trailing data: [DONE]
	// - markdown code fences around the JSON content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(proxyResponse(t, "```json\n{\"score\":88}\n```", "data: [DONE]"))
	}))
	defer server.Close()

	client := &AnthropicClient{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		http:    server.Client(),
	}

	t.Run("struct result", func(t *testing.T) {
		var result struct {
			Score int `json:"score"`
		}
		err := client.Chat(context.Background(), "system", "user", &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Score != 88 {
			t.Fatalf("expected 88, got %d", result.Score)
		}
	})

	t.Run("string result", func(t *testing.T) {
		var raw string
		err := client.Chat(context.Background(), "system", "user", &raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `{"score":88}`
		if raw != want {
			t.Fatalf("got %q, want %q", raw, want)
		}
	})
}

func TestOpenAIClientAnthropicProxyResponse(t *testing.T) {
	// Simulate the 9Router proxy returning Anthropic format via the OpenAI endpoint.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(proxyResponse(t, "```json\n{\"score\":77}\n```", "data: [DONE]"))
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		http:    server.Client(),
	}

	t.Run("struct result", func(t *testing.T) {
		var result struct {
			Score int `json:"score"`
		}
		err := client.Chat(context.Background(), "system", "user", &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Score != 77 {
			t.Fatalf("expected 77, got %d", result.Score)
		}
	})

	t.Run("string result", func(t *testing.T) {
		var raw string
		err := client.Chat(context.Background(), "system", "user", &raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `{"score":77}`
		if raw != want {
			t.Fatalf("got %q, want %q", raw, want)
		}
	})
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
