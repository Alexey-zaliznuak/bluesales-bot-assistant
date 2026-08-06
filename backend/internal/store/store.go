package store

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("не найдено")
	ErrConflict = errors.New("конфликт данных")
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Pool() *pgxpool.Pool { return s.pool }

type User struct {
	ID           string    `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Document struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Categories []string  `json:"categories"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type KBSnapshot struct {
	ID             string    `json:"id"`
	Content        string    `json:"-"`
	ContentHash    string    `json:"contentHash"`
	DocumentsCount int       `json:"documentsCount"`
	CharsCount     int       `json:"charsCount"`
	IsActive       bool      `json:"isActive"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Chat struct {
	ID           string    `json:"id"`
	UserID       string    `json:"-"`
	KBSnapshotID *string   `json:"kbSnapshotId"`
	Title        string    `json:"title"`
	Model        string    `json:"model"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Attachment struct {
	Filename string `json:"filename"`
	Size     int    `json:"size"`
	Content  string `json:"content,omitempty"`
}

type Usage struct {
	PromptTokens     int      `json:"promptTokens"`
	CompletionTokens int      `json:"completionTokens"`
	TotalTokens      int      `json:"totalTokens"`
	ReasoningTokens  int      `json:"reasoningTokens"`
	Cost             *float64 `json:"cost,omitempty"`
}

type Message struct {
	ID          string       `json:"id"`
	ChatID      string       `json:"chatId"`
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
	Usage       *Usage       `json:"usage"`
	Error       *string      `json:"error"`
	CreatedAt   time.Time    `json:"createdAt"`
}
