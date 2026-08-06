package httpapi

import (
	"strings"

	"github.com/alex/bluesales-bot-assistant/backend/internal/openrouter"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

const responseFormattingPrompt = "Правила форматирования ответа:\n" +
	"Любой JSON помещай в отдельный Markdown-блок с именем файла в открывающей строке:\n" +
	"```json filename=\"Краткое понятное имя.json\"\n...\n```\n" +
	"Не выводи имя файла отдельной строкой. Для каждого отдельного JSON используй отдельный блок."

func responseFormattingMessage() openrouter.Message {
	return openrouter.TextMessage("system", responseFormattingPrompt)
}

func (s *Server) knowledgeBaseMessage(content string) openrouter.Message {
	return openrouter.TextMessage("system", content)
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
