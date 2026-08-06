package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/knowledge"
	"github.com/alex/bluesales-bot-assistant/backend/internal/openrouter"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

type kbStatusResponse struct {
	Snapshot           *store.KBSnapshot `json:"snapshot"`
	DocumentsCount     int               `json:"documentsCount"`
	LastDocumentUpdate *time.Time        `json:"lastDocumentUpdate"`
	CurrentHash        string            `json:"currentHash"`
	CurrentCharsCount  int               `json:"currentCharsCount"`
	Stale              bool              `json:"stale"`
	Model              string            `json:"model"`
	CacheMode          string            `json:"cacheMode"`
	CacheTTL           string            `json:"cacheTtl"`
	ReasoningEffort    string            `json:"reasoningEffort"`
	OpenRouterKeySet   bool              `json:"openrouterKeySet"`
}

func (s *Server) handleKBStatus(w http.ResponseWriter, r *http.Request) {
	built, err := s.buildKnowledgeBase(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}

	count, lastUpdate, err := s.store.DocumentsStats(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}

	snapshot, err := s.store.ActiveSnapshot(r.Context())
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, kbStatusResponse{
		Snapshot:           snapshot,
		DocumentsCount:     count,
		LastDocumentUpdate: lastUpdate,
		CurrentHash:        built.Hash,
		CurrentCharsCount:  built.CharsCount,
		Stale:              snapshot == nil || snapshot.ContentHash != built.Hash,
		Model:              s.cfg.OpenRouter.Model,
		CacheMode:          s.cfg.OpenRouter.CacheMode,
		CacheTTL:           s.cfg.OpenRouter.CacheTTL,
		ReasoningEffort:    s.cfg.OpenRouter.ReasoningEffort,
		OpenRouterKeySet:   s.client.Configured(),
	})
}

func (s *Server) handleKBPreview(w http.ResponseWriter, r *http.Request) {
	built, err := s.buildKnowledgeBase(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"content":        built.Content,
		"hash":           built.Hash,
		"cacheKey":       built.CacheKey,
		"documentsCount": built.DocumentsCount,
		"charsCount":     built.CharsCount,
	})
}

type kbSyncResponse struct {
	Snapshot *store.KBSnapshot `json:"snapshot"`
	Warmed   bool              `json:"warmed"`
	WarmSkip string            `json:"warmSkipped,omitempty"`
	Error    string            `json:"warmError,omitempty"`
}

// handleKBSync пересобирает базу знаний, делает снимок активным и прогревает
// кэш префикса в OpenRouter, чтобы первый запрос в чате уже попал в кэш.
func (s *Server) handleKBSync(w http.ResponseWriter, r *http.Request) {
	built, err := s.buildKnowledgeBase(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if built.DocumentsCount == 0 {
		writeError(w, http.StatusBadRequest, "в базе знаний нет ни одного документа")
		return
	}

	snapshot, err := s.store.ActivateSnapshot(r.Context(), built.Content, built.Hash, built.CacheKey,
		built.DocumentsCount, built.CharsCount)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	resp := kbSyncResponse{Snapshot: snapshot}

	if !s.client.Configured() {
		resp.WarmSkip = "OPENROUTER_API_KEY не задан, кэш не прогрет"
		writeJSON(w, http.StatusOK, resp)
		return
	}

	// Прогрев не должен зависеть от того, ушёл ли клиент со страницы.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), s.cfg.OpenRouter.Timeout)
	defer cancel()

	usage, warmErr := s.warmCache(ctx, snapshot)
	now := time.Now()

	var (
		promptTokens *int
		cachedTokens *int
		warmErrText  *string
	)
	if usage != nil {
		p, c := usage.PromptTokens, usage.CachedTokens()
		promptTokens, cachedTokens = &p, &c
	}
	if warmErr != nil {
		text := warmErr.Error()
		warmErrText = &text
		resp.Error = text
		slog.Error("прогрев кэша базы знаний", "error", warmErr)
	} else {
		resp.Warmed = true
	}

	warmedAt := &now
	if warmErr != nil {
		warmedAt = nil
	}
	if err := s.store.SetSnapshotWarmResult(ctx, snapshot.ID, warmedAt, promptTokens, cachedTokens, warmErrText); err != nil {
		slog.Error("сохранение результата прогрева", "error", err)
	}

	updated, err := s.store.GetSnapshot(ctx, snapshot.ID)
	if err == nil {
		resp.Snapshot = updated
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) buildKnowledgeBase(ctx context.Context) (knowledge.BuildResult, error) {
	docs, err := s.store.ListDocumentsForKnowledgeBase(ctx)
	if err != nil {
		return knowledge.BuildResult{}, err
	}
	return knowledge.Build(docs), nil
}

// warmCache делает минимальный запрос с полным префиксом: провайдер запишет
// префикс в кэш, а последующие чаты будут читать его по сниженной цене.
func (s *Server) warmCache(ctx context.Context, snapshot *store.KBSnapshot) (*openrouter.Usage, error) {
	maxTokens := 16
	noReasoning := "none"

	req := openrouter.ChatRequest{
		Messages: []openrouter.Message{
			s.knowledgeBaseMessage(snapshot.Content),
			openrouter.TextMessage("user", "Ответь одним словом: готов"),
		},
		MaxTokens: &maxTokens,
		// Прогрев не должен оплачивать рассуждения: кэшируется префикс промпта,
		// а не генерация.
		Reasoning:      &openrouter.Reasoning{Effort: noReasoning},
		SessionID:      snapshot.CacheKey,
		PromptCacheKey: snapshot.CacheKey,
	}

	result, err := s.client.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	return result.Usage, nil
}
