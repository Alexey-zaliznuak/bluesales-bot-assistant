package httpapi

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

type documentRequest struct {
	Title      *string   `json:"title"`
	Categories *[]string `json:"categories"`
	Body       *string   `json:"body"`
}

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := s.store.ListDocuments(r.Context(), store.DocumentFilter{
		Category: strings.TrimSpace(r.URL.Query().Get("category")),
		Search:   strings.TrimSpace(r.URL.Query().Get("search")),
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.store.ListCategories(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	doc, err := s.store.GetDocument(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) handleCreateDocument(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	title := ""
	if req.Title != nil {
		title = strings.TrimSpace(*req.Title)
	}
	if title == "" {
		writeError(w, http.StatusBadRequest, "заголовок обязателен")
		return
	}

	categories := []string{}
	if req.Categories != nil {
		categories = normalizeCategories(*req.Categories)
	}
	body := ""
	if req.Body != nil {
		body = *req.Body
	}

	doc, err := s.store.CreateDocument(r.Context(), title, categories, body)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (s *Server) handleUpdateDocument(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	patch := store.DocumentPatch{Body: req.Body}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			writeError(w, http.StatusBadRequest, "заголовок не может быть пустым")
			return
		}
		patch.Title = &title
	}
	if req.Categories != nil {
		categories := normalizeCategories(*req.Categories)
		patch.Categories = &categories
	}

	doc, err := s.store.UpdateDocument(r.Context(), chi.URLParam(r, "id"), patch)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteDocument(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// normalizeCategories убирает пустые значения и дубли, сохраняя порядок ввода.
func normalizeCategories(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, c := range in {
		c = strings.TrimSpace(c)
		if c == "" || seen[strings.ToLower(c)] {
			continue
		}
		seen[strings.ToLower(c)] = true
		out = append(out, c)
	}
	return out
}
