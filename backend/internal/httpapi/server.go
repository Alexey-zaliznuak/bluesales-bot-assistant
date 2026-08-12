package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/alex/bluesales-bot-assistant/backend/internal/config"
	"github.com/alex/bluesales-bot-assistant/backend/internal/openrouter"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

type Server struct {
	cfg    *config.Config
	store  *store.Store
	client *openrouter.Client
}

func NewServer(cfg *config.Config, st *store.Store, client *openrouter.Client) *Server {
	return &Server{cfg: cfg, store: st, client: client}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		r.Post("/auth/login", s.handleLogin)
		r.Post("/auth/register", s.handleRegister)

		r.Group(func(r chi.Router) {
			r.Use(s.requireAuth)

			r.Get("/auth/me", s.handleMe)
			r.Post("/auth/logout", s.handleLogout)

			r.Get("/health/openrouter", s.handleOpenRouterHealth)

			r.Get("/documents", s.handleListDocuments)
			r.Post("/documents", s.handleCreateDocument)
			r.Get("/documents/categories", s.handleListCategories)
			r.Get("/documents/{id}", s.handleGetDocument)
			r.Patch("/documents/{id}", s.handleUpdateDocument)
			r.Delete("/documents/{id}", s.handleDeleteDocument)

			r.Get("/kb/status", s.handleKBStatus)
			r.Get("/kb/preview", s.handleKBPreview)
			r.Post("/kb/sync", s.handleKBSync)

			r.Get("/chats", s.handleListChats)
			r.Post("/chats", s.handleCreateChat)
			r.Get("/chats/{id}", s.handleGetChat)
			r.Patch("/chats/{id}", s.handleRenameChat)
			r.Delete("/chats/{id}", s.handleDeleteChat)
			r.Post("/chats/{id}/messages", s.handleSendMessage)
		})
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 3*time.Second)
	defer cancel()

	if err := s.store.Pool().Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "down", "database": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":            "ok",
		"model":             s.client.Model(),
		"openrouterKeySet":  s.client.Configured(),
		"openrouterProxied": s.cfg.OpenRouter.ProxyURL != "",
	})
}

func (s *Server) handleOpenRouterHealth(w http.ResponseWriter, r *http.Request) {
	if !s.client.Configured() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "down",
			"error":  "OPENROUTER_API_KEY не задан",
		})
		return
	}
	if err := s.client.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "down", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"model":   s.client.Model(),
		"proxied": s.cfg.OpenRouter.ProxyURL != "",
	})
}
