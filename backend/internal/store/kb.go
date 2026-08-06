package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

const kbSnapshotColumns = `id, content, content_hash, documents_count, chars_count, is_active, created_at`

func scanSnapshot(row pgx.Row) (*KBSnapshot, error) {
	var s KBSnapshot
	err := row.Scan(&s.ID, &s.Content, &s.ContentHash, &s.DocumentsCount, &s.CharsCount,
		&s.IsActive, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ActivateSnapshot сохраняет снимок и делает его единственным активным.
// Повторная синхронизация с тем же содержимым переиспользует существующую
// строку, чтобы не плодить дубли.
func (s *Store) ActivateSnapshot(ctx context.Context, content, contentHash string, documentsCount, charsCount int) (*KBSnapshot, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `UPDATE kb_snapshots SET is_active = false WHERE is_active`); err != nil {
		return nil, err
	}

	snapshot, err := scanSnapshot(tx.QueryRow(ctx, `
		INSERT INTO kb_snapshots (content, content_hash, documents_count, chars_count, is_active)
		VALUES ($1, $2, $3, $4, true)
		ON CONFLICT (content_hash) DO UPDATE
			SET is_active = true,
			    documents_count = EXCLUDED.documents_count,
			    chars_count = EXCLUDED.chars_count
		RETURNING `+kbSnapshotColumns, content, contentHash, documentsCount, charsCount))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *Store) ActiveSnapshot(ctx context.Context) (*KBSnapshot, error) {
	return scanSnapshot(s.pool.QueryRow(ctx,
		`SELECT `+kbSnapshotColumns+` FROM kb_snapshots WHERE is_active LIMIT 1`))
}

func (s *Store) GetSnapshot(ctx context.Context, id string) (*KBSnapshot, error) {
	return scanSnapshot(s.pool.QueryRow(ctx,
		`SELECT `+kbSnapshotColumns+` FROM kb_snapshots WHERE id = $1`, id))
}
