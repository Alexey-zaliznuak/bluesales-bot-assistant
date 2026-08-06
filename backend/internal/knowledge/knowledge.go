package knowledge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

// SystemInstructions — начало контекста базы знаний. Любое изменение текста
// меняет хеш и требует новой синхронизации.
const SystemInstructions = `Ты — ассистент по настройке и созданию ботов в сервисе BlueSales.

Твоя роль:
- помогаешь проектировать и настраивать автоворонки, сценарии диалогов, автоответы и рассылки в BlueSales;
- объясняешь, какие блоки, условия, теги и переменные нужно использовать и в каком порядке;
- по запросу выдаёшь готовую конфигурацию бота, повторяя структуру и терминологию из примеров ниже.

Правила работы:
1. База знаний ниже — единственный достоверный источник по возможностям и синтаксису BlueSales. Опирайся в первую очередь на неё.
2. Если в базе знаний нет ответа, прямо скажи об этом и обозначь, что даёшь общую рекомендацию, а не выдержку из документации. Не выдумывай названия полей, блоков и настроек.
3. Ссылайся на документы по их заголовкам, когда берёшь оттуда конкретную настройку.
4. Отвечай по-русски, конкретно и по шагам. Готовые конфигурации и сценарии оформляй так же, как в примерах базы знаний.
5. Если запрос неоднозначный, задай один-два уточняющих вопроса вместо догадок.`

// Format описывает разметку склейки и входит в хеш снимка.
const Format = "xml/v1"

type BuildResult struct {
	Content        string
	Hash           string
	DocumentsCount int
	CharsCount     int
}

// Build склеивает документы в один текст для системного сообщения.
// Стабильный порядок документов обеспечивает стабильный хеш снимка.
func Build(docs []store.Document) BuildResult {
	var sb strings.Builder

	sb.WriteString("<system_instructions>\n")
	sb.WriteString(SystemInstructions)
	sb.WriteString("\n</system_instructions>\n\n")

	fmt.Fprintf(&sb, "<knowledge_base format=%q documents=\"%d\">\n", Format, len(docs))
	for _, doc := range docs {
		fmt.Fprintf(&sb, "<document id=%q title=%q categories=%q>\n",
			escapeAttr(doc.ID),
			escapeAttr(doc.Title),
			escapeAttr(strings.Join(doc.Categories, ", ")),
		)
		sb.WriteString(escapeBody(doc.Body))
		sb.WriteString("\n</document>\n")
	}
	sb.WriteString("</knowledge_base>")

	content := sb.String()
	sum := sha256.Sum256([]byte(content))
	hash := hex.EncodeToString(sum[:])

	return BuildResult{
		Content:        content,
		Hash:           hash,
		DocumentsCount: len(docs),
		CharsCount:     len([]rune(content)),
	}
}

var attrEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"\n", " ",
	"\r", " ",
)

func escapeAttr(v string) string {
	return attrEscaper.Replace(v)
}

// escapeBody не экранирует разметку целиком — тело примеров настроек
// читабельнее без этого, — но закрывает возможность порвать контейнер
// документа изнутри.
var bodyEscaper = strings.NewReplacer(
	"</document>", "&lt;/document&gt;",
	"</knowledge_base>", "&lt;/knowledge_base&gt;",
)

func escapeBody(v string) string {
	return bodyEscaper.Replace(strings.TrimRight(v, "\n"))
}
