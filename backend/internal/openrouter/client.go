package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/config"
)

var ErrNoAPIKey = errors.New("OPENROUTER_API_KEY не задан")

type apiError struct {
	Code     any             `json:"code"`
	Message  string          `json:"message"`
	Metadata json.RawMessage `json:"metadata"`
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("OpenRouter %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("OpenRouter %d: %s", e.StatusCode, e.Message)
}

type Client struct {
	cfg  config.OpenRouter
	http *http.Client
}

func New(cfg config.OpenRouter) (*Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("разбор OPENROUTER_PROXY_URL: %w", err)
		}
		switch proxyURL.Scheme {
		case "http", "https", "socks5", "socks5h":
		default:
			return nil, fmt.Errorf("неподдерживаемая схема прокси %q (нужны http, https, socks5)", proxyURL.Scheme)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	} else {
		// Прокси не задан — идём напрямую, игнорируя системные HTTP_PROXY.
		transport.Proxy = nil
	}

	return &Client{
		cfg:  cfg,
		http: &http.Client{Transport: transport, Timeout: cfg.Timeout},
	}, nil
}

func (c *Client) Config() config.OpenRouter { return c.cfg }

func (c *Client) Model() string { return c.cfg.Model }

// Configured сообщает, можно ли вообще ходить в API.
func (c *Client) Configured() bool { return c.cfg.APIKey != "" }

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	if c.cfg.APIKey == "" {
		return nil, ErrNoAPIKey
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.HTTPReferer != "" {
		req.Header.Set("HTTP-Referer", c.cfg.HTTPReferer)
	}
	if c.cfg.AppTitle != "" {
		req.Header.Set("X-Title", c.cfg.AppTitle)
	}
	return req, nil
}

// ApplyDefaults проставляет модель, режим рассуждений и параметры кэширования
// префикса, если вызывающий код не задал их явно.
func (c *Client) ApplyDefaults(req *ChatRequest) {
	if req.Model == "" {
		req.Model = c.cfg.Model
	}
	if req.Reasoning == nil && c.cfg.ReasoningEffort != "" && c.cfg.ReasoningEffort != "none" {
		req.Reasoning = &Reasoning{Effort: c.cfg.ReasoningEffort}
	}
	if req.Usage == nil {
		req.Usage = &UsageOption{Include: true}
	}
	if req.PromptCacheOptions == nil && c.cfg.CacheMode == "explicit" {
		req.PromptCacheOptions = &PromptCacheOptions{Mode: "explicit", TTL: c.cfg.CacheTTL}
	}
}

func (c *Client) Complete(ctx context.Context, req ChatRequest) (*CompletionResult, error) {
	req.Stream = false
	req.StreamOptions = nil
	c.ApplyDefaults(&req)

	httpReq, err := c.newRequest(ctx, http.MethodPost, "/chat/completions", req)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("запрос к OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(resp.StatusCode, body)
	}

	var parsed completionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("разбор ответа OpenRouter: %w", err)
	}
	if parsed.Error != nil {
		return nil, &APIError{StatusCode: resp.StatusCode, Code: codeToString(parsed.Error.Code), Message: parsed.Error.Message}
	}

	result := &CompletionResult{Model: parsed.Model, Usage: parsed.Usage}
	if len(parsed.Choices) > 0 {
		result.Content = parsed.Choices[0].Message.Content
		result.Reasoning = parsed.Choices[0].Message.Reasoning
	}
	return result, nil
}

type StreamCallbacks struct {
	OnContent   func(string) error
	OnReasoning func(string) error
	OnUsage     func(Usage)
}

// StreamChat читает SSE-поток OpenRouter и отдаёт дельты в колбэки.
func (c *Client) StreamChat(ctx context.Context, req ChatRequest, cb StreamCallbacks) error {
	req.Stream = true
	if req.StreamOptions == nil {
		req.StreamOptions = &StreamOptions{IncludeUsage: true}
	}
	c.ApplyDefaults(&req)

	httpReq, err := c.newRequest(ctx, http.MethodPost, "/chat/completions", req)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("запрос к OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return parseHTTPError(resp.StatusCode, body)
	}

	reader := bufio.NewReaderSize(resp.Body, 64*1024)
	for {
		line, err := readSSELine(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		// Комментарии вида ": OPENROUTER PROCESSING" — keep-alive.
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		data, ok := strings.CutPrefix(line, "data:")
		if !ok {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			return nil
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// Битый кусок потока не повод ронять весь ответ.
			continue
		}
		if chunk.Error != nil {
			return &APIError{StatusCode: http.StatusBadGateway, Code: codeToString(chunk.Error.Code), Message: chunk.Error.Message}
		}

		for _, choice := range chunk.Choices {
			if choice.Delta.Reasoning != "" && cb.OnReasoning != nil {
				if err := cb.OnReasoning(choice.Delta.Reasoning); err != nil {
					return err
				}
			}
			if choice.Delta.Content != "" && cb.OnContent != nil {
				if err := cb.OnContent(choice.Delta.Content); err != nil {
					return err
				}
			}
		}

		if chunk.Usage != nil && cb.OnUsage != nil {
			cb.OnUsage(*chunk.Usage)
		}
	}
}

// Ping проверяет, что до OpenRouter есть сеть и ключ принят.
func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	req, err := c.newRequest(ctx, http.MethodGet, "/key", nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("нет связи с OpenRouter: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK {
		return parseHTTPError(resp.StatusCode, body)
	}
	return nil
}

// readSSELine читает строку произвольной длины: длинные дельты рассуждений
// не влезают в буфер bufio.Scanner по умолчанию.
func readSSELine(r *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		chunk, isPrefix, err := r.ReadLine()
		if err != nil {
			if sb.Len() > 0 && errors.Is(err, io.EOF) {
				return sb.String(), nil
			}
			return "", err
		}
		sb.Write(chunk)
		if !isPrefix {
			return sb.String(), nil
		}
	}
}

func parseHTTPError(status int, body []byte) error {
	var wrapper struct {
		Error *apiError `json:"error"`
	}
	if err := json.Unmarshal(body, &wrapper); err == nil && wrapper.Error != nil {
		return &APIError{StatusCode: status, Code: codeToString(wrapper.Error.Code), Message: wrapper.Error.Message}
	}
	message := strings.TrimSpace(string(body))
	if len(message) > 500 {
		message = message[:500]
	}
	if message == "" {
		message = http.StatusText(status)
	}
	return &APIError{StatusCode: status, Message: message}
}

func codeToString(code any) string {
	switch v := code.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return fmt.Sprint(v)
	}
}
