package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/config"
	"github.com/alex/bluesales-bot-assistant/backend/internal/db"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("конфигурация", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("подключение к БД", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		slog.Error("миграции", "error", err)
		os.Exit(1)
	}

	slog.Info("миграции завершены")
}
