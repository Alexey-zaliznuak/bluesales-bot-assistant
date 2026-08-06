package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, userID, tokenHash, expiresAt)
	return err
}

// GetUserBySessionToken возвращает владельца непросроченной сессии.
func (s *Store) GetUserBySessionToken(ctx context.Context, tokenHash string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.login, u.password_hash, u.created_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()`, tokenHash,
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *Store) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= now()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
