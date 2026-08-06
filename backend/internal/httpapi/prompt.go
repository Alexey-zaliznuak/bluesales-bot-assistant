package httpapi

import (
	"strings"

	"github.com/alex/bluesales-bot-assistant/backend/internal/openrouter"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

// knowledgeBaseMessage оборачивает префикс базы знаний в system-сообщение
// с маркером кэширования на блоке.
func (s *Server) knowledgeBaseMessage(content string) openrouter.Message {
	block := openrouter.TextBlock{Type: "text", Text: content}

	switch s.cfg.OpenRouter.CacheMode {
	case "off":
	case "auto":
		block.CacheControl = &openrouter.CacheControl{Type: "ephemeral", TTL: s.cfg.OpenRouter.CacheTTL}
	default:
		// Anthropic и Google понимают только cache_control, причём TTL из него
		// не переносится в prompt_cache_breakpoint. Для OpenAI GPT-5.6+
		// родной формат — prompt_cache_breakpoint.
		if usesCacheControl(s.cfg.OpenRouter.Model) {
			block.CacheControl = &openrouter.CacheControl{Type: "ephemeral", TTL: s.cfg.OpenRouter.CacheTTL}
		} else {
			block.PromptCacheBreakpoint = &openrouter.PromptCacheBreakpoint{Mode: "explicit"}
		}
	}

	return openrouter.Message{Role: "system", Content: []openrouter.TextBlock{block}}
}

func usesCacheControl(model string) bool {
	return strings.HasPrefix(model, "anthropic/") ||
		strings.HasPrefix(model, "google/") ||
		strings.HasPrefix(model, "qwen/") ||
		strings.HasPrefix(model, "deepseek/")
}

// userContent собирает текст пользователя вместе с текстом вложений.
func userContent(text string, attachments []store.Attachment) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimSpace(text))

	if len(attachments) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("<attachments>\n")
		for _, a := range attachments {
			sb.WriteString("<attachment filename=\"")
			sb.WriteString(escapeAttrValue(a.Filename))
			sb.WriteString("\">\n")
			sb.WriteString(strings.ReplaceAll(a.Content, "</attachment>", "&lt;/attachment&gt;"))
			sb.WriteString("\n</attachment>\n")
		}
		sb.WriteString("</attachments>")
	}

	return sb.String()
}

var attrValueEscaper = strings.NewReplacer(`"`, "&quot;", "<", "&lt;", ">", "&gt;", "&", "&amp;", "\n", " ", "\r", " ")

func escapeAttrValue(v string) string { return attrValueEscaper.Replace(v) }
