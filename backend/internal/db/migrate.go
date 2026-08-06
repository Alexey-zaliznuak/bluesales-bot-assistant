package db

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alex/bluesales-bot-assistant/backend/migrations"
)

// Migrate применяет ещё не применённые .sql-файлы в лексикографическом порядке.
// Каждый файл выполняется в отдельной транзакции вместе с записью в журнал,
// поэтому частично применённая миграция откатывается целиком.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("создание schema_migrations: %w", err)
	}

	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("чтение schema_migrations: %w", err)
	}
	applied := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	entries, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(entries)

	for _, name := range entries {
		if applied[name] {
			continue
		}

		body, err := migrations.FS.ReadFile(name)
		if err != nil {
			return err
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("миграция %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("журнал миграции %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("коммит миграции %s: %w", name, err)
		}

		slog.Info("миграция применена", "version", name)
	}

	return nil
}
