package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/config"
	"github.com/alex/bluesales-bot-assistant/backend/internal/db"
	"github.com/alex/bluesales-bot-assistant/backend/internal/httpapi"
	"github.com/alex/bluesales-bot-assistant/backend/internal/openrouter"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("конфигурация", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connectCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	pool, err := db.Connect(connectCtx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		slog.Error("подключение к БД", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	client, err := openrouter.New(cfg.OpenRouter)
	if err != nil {
		slog.Error("клиент OpenRouter", "error", err)
		os.Exit(1)
	}
	if !client.Configured() {
		slog.Warn("OPENROUTER_API_KEY не задан: чаты работать не будут")
	}

	st := store.New(pool)
	go cleanupSessions(ctx, st)

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           httpapi.NewServer(cfg, st, client).Router(),
		ReadHeaderTimeout: 15 * time.Second,
		// WriteTimeout не выставляем: SSE-ответы живут минутами.
		IdleTimeout: 120 * time.Second,
	}

	go func() {
		slog.Info("API запущен",
			"port", cfg.APIPort,
			"model", cfg.OpenRouter.Model,
			"reasoningEffort", cfg.OpenRouter.ReasoningEffort,
			"proxy", cfg.OpenRouter.ProxyURL != "")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http-сервер", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("остановка сервера")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown", "error", err)
	}
}

func cleanupSessions(ctx context.Context, st *store.Store) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		if n, err := st.DeleteExpiredSessions(ctx); err != nil {
			if ctx.Err() == nil {
				slog.Error("очистка сессий", "error", err)
			}
		} else if n > 0 {
			slog.Info("удалены истёкшие сессии", "count", n)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
