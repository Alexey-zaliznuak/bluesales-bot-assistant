package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/auth"
	"github.com/alex/bluesales-bot-assistant/backend/internal/config"
	"github.com/alex/bluesales-bot-assistant/backend/internal/db"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("конфигурация", "error", err)
		os.Exit(1)
	}

	if cfg.SeedUserLogin == "" || cfg.SeedUserPassword == "" {
		slog.Error("SEED_USER_LOGIN и SEED_USER_PASSWORD обязательны для сида")
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

	hash, err := auth.HashPassword(cfg.SeedUserPassword)
	if err != nil {
		slog.Error("хеширование пароля", "error", err)
		os.Exit(1)
	}

	s := store.New(pool)
	user, created, err := s.UpsertUser(ctx, cfg.SeedUserLogin, hash)
	if err != nil {
		slog.Error("создание пользователя", "error", err)
		os.Exit(1)
	}

	if created {
		slog.Info("пользователь создан", "login", user.Login, "id", user.ID)
	} else {
		slog.Info("пользователь уже существовал, пароль обновлён", "login", user.Login, "id", user.ID)
	}
}
