package openrouter

// Content блока сообщения. Массив блоков нужен там, где на конкретный блок
// вешается маркер кэширования; для обычных сообщений Content — просто строка.
type TextBlock struct {
	Type                  string                 `json:"type"`
	Text                  string                 `json:"text"`
	CacheControl          *CacheControl          `json:"cache_control,omitempty"`
	PromptCacheBreakpoint *PromptCacheBreakpoint `json:"prompt_cache_breakpoint,omitempty"`
}

// CacheControl — формат Anthropic/Google. OpenRouter конвертирует его
// в prompt_cache_breakpoint, если запрос уходит в OpenAI.
type CacheControl struct {
	Type string `json:"type"`
	TTL  string `json:"ttl,omitempty"`
}

// PromptCacheBreakpoint — формат OpenAI GPT-5.6+: помечает конец
// переиспользуемого префикса промпта.
type PromptCacheBreakpoint struct {
	Mode string `json:"mode"`
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func TextMessage(role, text string) Message {
	return Message{Role: role, Content: text}
}

type Reasoning struct {
	Effort  string `json:"effort,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
	Exclude *bool  `json:"exclude,omitempty"`
}

type PromptCacheOptions struct {
	Mode string `json:"mode,omitempty"`
	TTL  string `json:"ttl,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type UsageOption struct {
	Include bool `json:"include"`
}

type ChatRequest struct {
	Model              string              `json:"model"`
	Messages           []Message           `json:"messages"`
	Stream             bool                `json:"stream,omitempty"`
	StreamOptions      *StreamOptions      `json:"stream_options,omitempty"`
	Reasoning          *Reasoning          `json:"reasoning,omitempty"`
	MaxTokens          *int                `json:"max_tokens,omitempty"`
	Temperature        *float64            `json:"temperature,omitempty"`
	Usage              *UsageOption        `json:"usage,omitempty"`
	SessionID          string              `json:"session_id,omitempty"`
	PromptCacheKey     string              `json:"prompt_cache_key,omitempty"`
	PromptCacheOptions *PromptCacheOptions `json:"prompt_cache_options,omitempty"`
}

type PromptTokensDetails struct {
	CachedTokens     int `json:"cached_tokens"`
	CacheWriteTokens int `json:"cache_write_tokens"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type Usage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	Cost                    *float64                 `json:"cost"`
	PromptTokensDetails     *PromptTokensDetails     `json:"prompt_tokens_details"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details"`
}

func (u Usage) CachedTokens() int {
	if u.PromptTokensDetails == nil {
		return 0
	}
	return u.PromptTokensDetails.CachedTokens
}

func (u Usage) ReasoningTokens() int {
	if u.CompletionTokensDetails == nil {
		return 0
	}
	return u.CompletionTokensDetails.ReasoningTokens
}

type choiceDelta struct {
	Content   string `json:"content"`
	Reasoning string `json:"reasoning"`
}

type choiceMessage struct {
	Content   string `json:"content"`
	Reasoning string `json:"reasoning"`
}

type streamChunk struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Delta        choiceDelta `json:"delta"`
		FinishReason *string     `json:"finish_reason"`
	} `json:"choices"`
	Usage *Usage    `json:"usage"`
	Error *apiError `json:"error"`
}

type completionResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message      choiceMessage `json:"message"`
		FinishReason *string       `json:"finish_reason"`
	} `json:"choices"`
	Usage *Usage    `json:"usage"`
	Error *apiError `json:"error"`
}

type CompletionResult struct {
	Content   string
	Reasoning string
	Model     string
	Usage     *Usage
}
