package store

import (
	"context"
	"time"
)

type DailyTokenUsage struct {
	Date        string `json:"date"`
	TotalTokens int64  `json:"totalTokens"`
}

type UserTokenUsage struct {
	ID            string    `json:"id"`
	Login         string    `json:"login"`
	CreatedAt     time.Time `json:"createdAt"`
	MonthTokens   int64     `json:"monthTokens"`
	AllTimeTokens int64     `json:"allTimeTokens"`
}

func (s *Store) DailyTokenUsage(ctx context.Context, start, end time.Time) ([]DailyTokenUsage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			to_char(m.created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD') AS day,
			COALESCE(SUM((m.usage->>'totalTokens')::bigint), 0)::bigint AS total_tokens
		FROM messages m
		WHERE m.usage IS NOT NULL
		  AND m.created_at >= $1
		  AND m.created_at < $2
		GROUP BY day
		ORDER BY day`,
		start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []DailyTokenUsage{}
	for rows.Next() {
		var item DailyTokenUsage
		if err := rows.Scan(&item.Date, &item.TotalTokens); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UserTokenUsage(ctx context.Context, monthStart, monthEnd time.Time) ([]UserTokenUsage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			u.id,
			u.login,
			u.created_at,
			COALESCE(
				SUM((m.usage->>'totalTokens')::bigint)
					FILTER (WHERE m.created_at >= $1 AND m.created_at < $2),
				0
			)::bigint AS month_tokens,
			COALESCE(SUM((m.usage->>'totalTokens')::bigint), 0)::bigint AS all_time_tokens
		FROM users u
		LEFT JOIN chats c ON c.user_id = u.id
		LEFT JOIN messages m ON m.chat_id = c.id AND m.usage IS NOT NULL
		GROUP BY u.id, u.login, u.created_at
		ORDER BY month_tokens DESC, all_time_tokens DESC, lower(u.login)`,
		monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []UserTokenUsage{}
	for rows.Next() {
		var item UserTokenUsage
		if err := rows.Scan(
			&item.ID,
			&item.Login,
			&item.CreatedAt,
			&item.MonthTokens,
			&item.AllTimeTokens,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
