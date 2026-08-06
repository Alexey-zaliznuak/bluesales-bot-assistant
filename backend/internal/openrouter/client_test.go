package openrouter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/config"
)

func testConfig(baseURL string) config.OpenRouter {
	return config.OpenRouter{
		APIKey:          "test-key",
		BaseURL:         baseURL,
		Model:           "openai/gpt-5.6-luna",
		ReasoningEffort: "max",
		Timeout:         10 * time.Second,
		AppTitle:        "test",
	}
}

func TestStreamChatParsesSSE(t *testing.T) {
	var captured map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)

		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q", got)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)

		for _, line := range []string{
			": OPENROUTER PROCESSING",
			`data: {"choices":[{"delta":{"reasoning":"думаю"}}]}`,
			`data: {"choices":[{"delta":{"content":"При"}}]}`,
			`data: {"choices":[{"delta":{"content":"вет"}}]}`,
			`data: {"choices":[],"usage":{"prompt_tokens":1200,"completion_tokens":8,"total_tokens":1208,` +
				`"completion_tokens_details":{"reasoning_tokens":4}}}`,
			"data: [DONE]",
		} {
			_, _ = io.WriteString(w, line+"\n\n")
			flusher.Flush()
		}
	}))
	defer server.Close()

	client, err := New(testConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	var content, reasoning strings.Builder
	var usage *Usage

	err = client.StreamChat(context.Background(), ChatRequest{
		Messages: []Message{
			TextMessage("system", "база знаний"),
			TextMessage("user", "привет"),
		},
	}, StreamCallbacks{
		OnContent:   func(d string) error { content.WriteString(d); return nil },
		OnReasoning: func(d string) error { reasoning.WriteString(d); return nil },
		OnUsage:     func(u Usage) { usage = &u },
	})
	if err != nil {
		t.Fatalf("StreamChat: %v", err)
	}

	if content.String() != "Привет" {
		t.Errorf("content = %q", content.String())
	}
	if reasoning.String() != "думаю" {
		t.Errorf("reasoning = %q", reasoning.String())
	}
	if usage == nil || usage.ReasoningTokens() != 4 {
		t.Fatalf("usage разобран неверно: %+v", usage)
	}

	// Настройки по умолчанию должны доехать до провайдера.
	if captured["model"] != "openai/gpt-5.6-luna" {
		t.Errorf("model = %v", captured["model"])
	}
	if effort := captured["reasoning"].(map[string]any)["effort"]; effort != "max" {
		t.Errorf("reasoning.effort = %v", effort)
	}
	if captured["stream_options"].(map[string]any)["include_usage"] != true {
		t.Error("include_usage не запрошен, usage не придёт в потоке")
	}
	if _, ok := captured["prompt_cache_options"]; ok {
		t.Error("параметры кэширования не должны отправляться")
	}
}

func TestStreamChatReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":{"code":429,"message":"rate limited"}}`)
	}))
	defer server.Close()

	client, err := New(testConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	err = client.StreamChat(context.Background(), ChatRequest{Messages: []Message{TextMessage("user", "hi")}}, StreamCallbacks{})
	if err == nil {
		t.Fatal("ожидалась ошибка")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("ожидался *APIError, получено %T", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests || apiErr.Message != "rate limited" {
		t.Errorf("неожиданная ошибка: %v", apiErr)
	}
}

// Ошибка может прийти внутри уже открытого потока.
func TestStreamChatReturnsInStreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, `data: {"error":{"code":"server_error","message":"провайдер упал"}}`+"\n\n")
	}))
	defer server.Close()

	client, err := New(testConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	err = client.StreamChat(context.Background(), ChatRequest{Messages: []Message{TextMessage("user", "hi")}}, StreamCallbacks{})
	if err == nil || !strings.Contains(err.Error(), "провайдер упал") {
		t.Fatalf("ошибка из потока потеряна: %v", err)
	}
}

func TestNewRejectsUnknownProxyScheme(t *testing.T) {
	cfg := testConfig("https://example.com")
	cfg.ProxyURL = "ftp://user:pass@host:1080"

	if _, err := New(cfg); err == nil {
		t.Fatal("ожидалась ошибка на неподдерживаемой схеме прокси")
	}
}

func TestStreamChatWithoutAPIKey(t *testing.T) {
	cfg := testConfig("https://example.com")
	cfg.APIKey = ""

	client, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if client.Configured() {
		t.Fatal("клиент без ключа не должен считаться настроенным")
	}
	if err := client.StreamChat(context.Background(), ChatRequest{}, StreamCallbacks{}); err != ErrNoAPIKey {
		t.Fatalf("ожидался ErrNoAPIKey, получено %v", err)
	}
}
