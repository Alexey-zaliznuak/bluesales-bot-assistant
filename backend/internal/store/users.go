package store

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *Store) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE lower(login) = lower($1)`, strings.TrimSpace(login),
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE id = $1`, id,
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateUser создаёт нового пользователя, не изменяя существующую учётную запись.
func (s *Store) CreateUser(ctx context.Context, login, passwordHash string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (lower(login)) DO NOTHING
		RETURNING id, login, password_hash, created_at`,
		strings.TrimSpace(login), passwordHash,
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConflict
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpsertUser создаёт пользователя или обновляет пароль существующего.
// Используется сид-скриптом, чтобы повторный запуск был идемпотентным.
func (s *Store) UpsertUser(ctx context.Context, login, passwordHash string) (*User, bool, error) {
	var (
		u       User
		created bool
	)
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (lower(login)) DO UPDATE
			SET password_hash = EXCLUDED.password_hash,
			    updated_at = now()
		RETURNING id, login, password_hash, created_at, (xmax = 0) AS created`,
		strings.TrimSpace(login), passwordHash,
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt, &created)
	if err != nil {
		return nil, false, err
	}
	return &u, created, nil
}
