package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateChat(ctx context.Context, userID, title, model string, snapshotID *string) (*Chat, error) {
	var c Chat
	err := s.pool.QueryRow(ctx, `
		INSERT INTO chats (user_id, title, model, kb_snapshot_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, kb_snapshot_id, title, model, created_at, updated_at`,
		userID, title, model, snapshotID,
	).Scan(&c.ID, &c.UserID, &c.KBSnapshotID, &c.Title, &c.Model, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListChats(ctx context.Context, userID string) ([]Chat, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, kb_snapshot_id, title, model, created_at, updated_at
		FROM chats WHERE user_id = $1
		ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := []Chat{}
	for rows.Next() {
		var c Chat
		if err := rows.Scan(&c.ID, &c.UserID, &c.KBSnapshotID, &c.Title, &c.Model, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}
	return chats, rows.Err()
}

func (s *Store) GetChat(ctx context.Context, userID, chatID string) (*Chat, error) {
	var c Chat
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, kb_snapshot_id, title, model, created_at, updated_at
		FROM chats WHERE id = $1 AND user_id = $2`, chatID, userID,
	).Scan(&c.ID, &c.UserID, &c.KBSnapshotID, &c.Title, &c.Model, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) RenameChat(ctx context.Context, userID, chatID, title string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE chats SET title = $3, updated_at = now() WHERE id = $1 AND user_id = $2`,
		chatID, userID, title)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) TouchChat(ctx context.Context, userID, chatID string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE chats SET updated_at = now() WHERE id = $1 AND user_id = $2`,
		chatID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteChat(ctx context.Context, userID, chatID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM chats WHERE id = $1 AND user_id = $2`, chatID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListMessages(ctx context.Context, userID, chatID string) ([]Message, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.id, m.chat_id, m.role, m.content, m.attachments, m.usage, m.error, m.created_at
		FROM messages m
		JOIN chats c ON c.id = m.chat_id
		WHERE m.chat_id = $1 AND c.user_id = $2
		ORDER BY m.seq`, chatID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var (
			m           Message
			attachments []byte
			usage       []byte
		)
		if err := rows.Scan(&m.ID, &m.ChatID, &m.Role, &m.Content, &attachments, &usage, &m.Error, &m.CreatedAt); err != nil {
			return nil, err
		}
		if len(attachments) > 0 {
			if err := json.Unmarshal(attachments, &m.Attachments); err != nil {
				return nil, err
			}
		}
		if m.Attachments == nil {
			m.Attachments = []Attachment{}
		}
		if len(usage) > 0 {
			if err := json.Unmarshal(usage, &m.Usage); err != nil {
				return nil, err
			}
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (s *Store) CreateMessage(ctx context.Context, userID, chatID, role, content string, attachments []Attachment, usage *Usage, msgErr *string) (*Message, error) {
	if attachments == nil {
		attachments = []Attachment{}
	}
	attachmentsJSON, err := json.Marshal(attachments)
	if err != nil {
		return nil, err
	}
	var usageJSON []byte
	if usage != nil {
		if usageJSON, err = json.Marshal(usage); err != nil {
			return nil, err
		}
	}

	var m Message
	var attachmentsOut, usageOut []byte
	err = s.pool.QueryRow(ctx, `
		INSERT INTO messages (chat_id, role, content, attachments, usage, error)
		SELECT c.id, $3, $4, $5, $6, $7
		FROM chats c
		WHERE c.id = $1 AND c.user_id = $2
		RETURNING id, chat_id, role, content, attachments, usage, error, created_at`,
		chatID, userID, role, content, attachmentsJSON, usageJSON, msgErr,
	).Scan(&m.ID, &m.ChatID, &m.Role, &m.Content, &attachmentsOut, &usageOut, &m.Error, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	m.Attachments = attachments
	m.Usage = usage
	return &m, nil
}

func (s *Store) DeleteMessage(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM messages m
		USING chats c
		WHERE m.id = $1 AND m.chat_id = c.id AND c.user_id = $2`,
		id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
