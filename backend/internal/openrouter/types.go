package openrouter

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

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type UsageOption struct {
	Include bool `json:"include"`
}

type ChatRequest struct {
	Model         string         `json:"model"`
	Messages      []Message      `json:"messages"`
	Stream        bool           `json:"stream,omitempty"`
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
	Reasoning     *Reasoning     `json:"reasoning,omitempty"`
	MaxTokens     *int           `json:"max_tokens,omitempty"`
	Temperature   *float64       `json:"temperature,omitempty"`
	Usage         *UsageOption   `json:"usage,omitempty"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type Usage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	Cost                    *float64                 `json:"cost"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details"`
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
