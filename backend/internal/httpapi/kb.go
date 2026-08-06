package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/knowledge"
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
		"documentsCount": built.DocumentsCount,
		"charsCount":     built.CharsCount,
	})
}

type kbSyncResponse struct {
	Snapshot *store.KBSnapshot `json:"snapshot"`
}

// handleKBSync пересобирает базу знаний и делает новый снимок активным.
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

	snapshot, err := s.store.ActivateSnapshot(r.Context(), built.Content, built.Hash,
		built.DocumentsCount, built.CharsCount)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, kbSyncResponse{Snapshot: snapshot})
}

func (s *Server) buildKnowledgeBase(ctx context.Context) (knowledge.BuildResult, error) {
	docs, err := s.store.ListDocumentsForKnowledgeBase(ctx)
	if err != nil {
		return knowledge.BuildResult{}, err
	}
	return knowledge.Build(docs), nil
}
