package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type DocumentFilter struct {
	Category string
	Search   string
}

func (s *Store) ListDocuments(ctx context.Context, f DocumentFilter) ([]Document, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, title, categories, body, created_at, updated_at
		FROM documents
		WHERE ($1 = '' OR $1 = ANY (categories))
		  AND ($2 = '' OR title ILIKE '%' || $2 || '%' OR body ILIKE '%' || $2 || '%')
		ORDER BY updated_at DESC`, f.Category, f.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := []Document{}
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.Title, &d.Categories, &d.Body, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// ListDocumentsForKnowledgeBase отдаёт документы в стабильном порядке:
// склеенный текст должен быть побайтно одинаковым между синхронизациями,
// иначе кэш префикса в OpenRouter не попадёт.
func (s *Store) ListDocumentsForKnowledgeBase(ctx context.Context) ([]Document, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, title, categories, body, created_at, updated_at
		FROM documents
		ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := []Document{}
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.Title, &d.Categories, &d.Body, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *Store) ListCategories(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT unnest(categories) AS category
		FROM documents
		ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetDocument(ctx context.Context, id string) (*Document, error) {
	var d Document
	err := s.pool.QueryRow(ctx, `
		SELECT id, title, categories, body, created_at, updated_at
		FROM documents WHERE id = $1`, id,
	).Scan(&d.ID, &d.Title, &d.Categories, &d.Body, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) CreateDocument(ctx context.Context, title string, categories []string, body string) (*Document, error) {
	if categories == nil {
		categories = []string{}
	}
	var d Document
	err := s.pool.QueryRow(ctx, `
		INSERT INTO documents (title, categories, body)
		VALUES ($1, $2, $3)
		RETURNING id, title, categories, body, created_at, updated_at`,
		title, categories, body,
	).Scan(&d.ID, &d.Title, &d.Categories, &d.Body, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

type DocumentPatch struct {
	Title      *string
	Categories *[]string
	Body       *string
}

func (s *Store) UpdateDocument(ctx context.Context, id string, p DocumentPatch) (*Document, error) {
	var categories any
	if p.Categories != nil {
		categories = *p.Categories
	}

	var d Document
	err := s.pool.QueryRow(ctx, `
		UPDATE documents SET
			title      = COALESCE($2, title),
			categories = COALESCE($3::text[], categories),
			body       = COALESCE($4, body),
			updated_at = now()
		WHERE id = $1
		RETURNING id, title, categories, body, created_at, updated_at`,
		id, p.Title, categories, p.Body,
	).Scan(&d.ID, &d.Title, &d.Categories, &d.Body, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) DeleteDocument(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM documents WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DocumentsStats нужен, чтобы показать «база знаний устарела» без пересборки текста.
func (s *Store) DocumentsStats(ctx context.Context) (count int, lastUpdate *time.Time, err error) {
	err = s.pool.QueryRow(ctx, `SELECT count(*), max(updated_at) FROM documents`).Scan(&count, &lastUpdate)
	return count, lastUpdate, err
}
