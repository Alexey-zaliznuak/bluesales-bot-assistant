package httpapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"

	"github.com/alex/bluesales-bot-assistant/backend/internal/openrouter"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

const defaultChatTitle = "Новый чат"

type chatResponse struct {
	*store.Chat
	KnowledgeBase *knowledgeBaseInfo `json:"knowledgeBase"`
	Messages      []store.Message    `json:"messages"`
}

type knowledgeBaseInfo struct {
	SnapshotID     string    `json:"snapshotId"`
	DocumentsCount int       `json:"documentsCount"`
	CharsCount     int       `json:"charsCount"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (s *Server) handleListChats(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	chats, err := s.store.ListChats(r.Context(), user.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chats)
}

type createChatRequest struct {
	Title string `json:"title"`
}

// handleCreateChat фиксирует за чатом активный снимок базы знаний: пересборка
// базы позже не должна менять контекст уже начатого диалога.
func (s *Server) handleCreateChat(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	var req createChatRequest
	if r.ContentLength > 0 {
		if !decodeJSON(w, r, &req) {
			return
		}
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = defaultChatTitle
	}

	var snapshotID *string
	snapshot, err := s.store.ActiveSnapshot(r.Context())
	switch {
	case err == nil:
		snapshotID = &snapshot.ID
	case errors.Is(err, store.ErrNotFound):
	default:
		writeStoreError(w, err)
		return
	}

	chat, err := s.store.CreateChat(r.Context(), user.ID, title, s.cfg.OpenRouter.Model, snapshotID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, chatResponse{
		Chat:          chat,
		KnowledgeBase: snapshotInfo(snapshot),
		Messages:      []store.Message{},
	})
}

func (s *Server) handleGetChat(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	chat, err := s.store.GetChat(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	messages, err := s.store.ListMessages(r.Context(), user.ID, chat.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	var snapshot *store.KBSnapshot
	if chat.KBSnapshotID != nil {
		if snapshot, err = s.store.GetSnapshot(r.Context(), *chat.KBSnapshotID); err != nil && !errors.Is(err, store.ErrNotFound) {
			writeStoreError(w, err)
			return
		}
	}

	writeJSON(w, http.StatusOK, chatResponse{
		Chat:          chat,
		KnowledgeBase: snapshotInfo(snapshot),
		Messages:      messages,
	})
}

func (s *Server) handleRenameChat(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	var req createChatRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "название не может быть пустым")
		return
	}

	if err := s.store.RenameChat(r.Context(), user.ID, chi.URLParam(r, "id"), title); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeleteChat(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	if err := s.store.DeleteChat(r.Context(), user.ID, chi.URLParam(r, "id")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleSendMessage принимает multipart (текст + текстовые файлы) и отдаёт
// ответ модели SSE-потоком.
func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	chat, err := s.store.GetChat(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	if !s.client.Configured() {
		writeError(w, http.StatusServiceUnavailable, "OPENROUTER_API_KEY не задан")
		return
	}

	maxBody := s.cfg.Upload.MaxSizeBytes*int64(s.cfg.Upload.MaxFiles) + 1<<20
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "не удалось прочитать форму: "+err.Error())
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	text := strings.TrimSpace(r.FormValue("content"))
	attachments, err := s.readAttachments(r.MultipartForm)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if text == "" && len(attachments) == 0 {
		writeError(w, http.StatusBadRequest, "пустое сообщение")
		return
	}

	history, err := s.store.ListMessages(r.Context(), user.ID, chat.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	snapshot, err := s.resolveSnapshot(r.Context(), chat)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	userMessage, err := s.store.CreateMessage(r.Context(), user.ID, chat.ID, "user", text, attachments, nil, nil)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	if chat.Title == defaultChatTitle && len(history) == 0 {
		if title := deriveTitle(text, attachments); title != "" {
			if err := s.store.RenameChat(r.Context(), user.ID, chat.ID, title); err != nil {
				slog.Error("автоименование чата", "error", err)
			} else {
				chat.Title = title
			}
		}
	}

	messages := make([]openrouter.Message, 0, len(history)+3)
	messages = append(messages, responseFormattingMessage())
	if snapshot != nil {
		messages = append(messages, s.knowledgeBaseMessage(snapshot.Content))
	}
	for _, m := range history {
		if m.Role == "assistant" && m.Content == "" {
			continue
		}
		messages = append(messages, openrouter.TextMessage(m.Role, userContent(m.Content, m.Attachments)))
	}
	messages = append(messages, openrouter.TextMessage("user", userContent(text, attachments)))

	req := openrouter.ChatRequest{Messages: messages}

	s.streamAnswer(w, r, chat, userMessage, req)
}

func (s *Server) streamAnswer(w http.ResponseWriter, r *http.Request, chat *store.Chat, userMessage *store.Message, req openrouter.ChatRequest) {
	sse := newSSEWriter(w)
	if err := sse.send("user_message", userMessage); err != nil {
		return
	}
	if chat.Title != "" {
		_ = sse.send("chat", map[string]string{"id": chat.ID, "title": chat.Title})
	}

	var (
		content   strings.Builder
		reasoning strings.Builder
		usage     *openrouter.Usage
	)

	streamErr := s.client.StreamChat(r.Context(), req, openrouter.StreamCallbacks{
		OnContent: func(delta string) error {
			content.WriteString(delta)
			return sse.send("delta", map[string]string{"delta": delta})
		},
		OnReasoning: func(delta string) error {
			reasoning.WriteString(delta)
			return sse.send("reasoning", map[string]string{"delta": delta})
		},
		OnUsage: func(u openrouter.Usage) {
			usage = &u
			_ = sse.send("usage", toStoreUsage(&u))
		},
	})

	// Клиент мог отключиться — сообщение всё равно должно сохраниться.
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 15*time.Second)
	defer cancel()

	var errText *string
	if streamErr != nil && !errors.Is(streamErr, context.Canceled) {
		text := humanizeStreamError(streamErr)
		errText = &text
		slog.Error("стриминг ответа", "chat", chat.ID, "error", streamErr)
	}

	if content.Len() == 0 && errText == nil && streamErr != nil {
		// Прервано клиентом без единого токена — сохранять нечего.
		return
	}

	assistantMessage, err := s.store.CreateMessage(saveCtx, chat.UserID, chat.ID, "assistant", content.String(), nil, toStoreUsage(usage), errText)
	if err != nil {
		slog.Error("сохранение ответа", "error", err)
		_ = sse.send("error", map[string]string{"error": "ответ получен, но не сохранён"})
		return
	}
	if err := s.store.TouchChat(saveCtx, chat.UserID, chat.ID); err != nil {
		slog.Error("обновление чата", "error", err)
	}

	if errText != nil {
		_ = sse.send("error", map[string]string{"error": *errText})
	}
	_ = sse.send("done", assistantMessage)
}

func (s *Server) resolveSnapshot(ctx context.Context, chat *store.Chat) (*store.KBSnapshot, error) {
	if chat.KBSnapshotID != nil {
		snapshot, err := s.store.GetSnapshot(ctx, *chat.KBSnapshotID)
		if err == nil {
			return snapshot, nil
		}
		if !errors.Is(err, store.ErrNotFound) {
			return nil, err
		}
	}

	snapshot, err := s.store.ActiveSnapshot(ctx)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	return snapshot, err
}

func (s *Server) readAttachments(form *multipart.Form) ([]store.Attachment, error) {
	if form == nil {
		return nil, nil
	}
	files := form.File["files"]
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > s.cfg.Upload.MaxFiles {
		return nil, fmt.Errorf("слишком много файлов: максимум %d", s.cfg.Upload.MaxFiles)
	}

	attachments := make([]store.Attachment, 0, len(files))
	for _, header := range files {
		name := filepath.Base(header.Filename)
		ext := strings.ToLower(filepath.Ext(name))
		if !slices.Contains(s.cfg.Upload.AllowedExtensions, ext) {
			return nil, fmt.Errorf("файл %s: поддерживаются только текстовые форматы (%s)",
				name, strings.Join(s.cfg.Upload.AllowedExtensions, ", "))
		}
		if header.Size > s.cfg.Upload.MaxSizeBytes {
			return nil, fmt.Errorf("файл %s больше %d МБ", name, s.cfg.Upload.MaxSizeBytes/(1024*1024))
		}

		file, err := header.Open()
		if err != nil {
			return nil, fmt.Errorf("файл %s: %w", name, err)
		}
		data, err := io.ReadAll(io.LimitReader(file, s.cfg.Upload.MaxSizeBytes+1))
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("файл %s: %w", name, err)
		}
		if int64(len(data)) > s.cfg.Upload.MaxSizeBytes {
			return nil, fmt.Errorf("файл %s больше %d МБ", name, s.cfg.Upload.MaxSizeBytes/(1024*1024))
		}

		text := strings.TrimPrefix(string(data), "\ufeff")
		if !utf8.ValidString(text) {
			return nil, fmt.Errorf("файл %s не является текстом в UTF-8", name)
		}

		attachments = append(attachments, store.Attachment{
			Filename: name,
			Size:     len(data),
			Content:  text,
		})
	}
	return attachments, nil
}

func snapshotInfo(snapshot *store.KBSnapshot) *knowledgeBaseInfo {
	if snapshot == nil {
		return nil
	}
	return &knowledgeBaseInfo{
		SnapshotID:     snapshot.ID,
		DocumentsCount: snapshot.DocumentsCount,
		CharsCount:     snapshot.CharsCount,
		CreatedAt:      snapshot.CreatedAt,
	}
}

func toStoreUsage(u *openrouter.Usage) *store.Usage {
	if u == nil {
		return nil
	}
	return &store.Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
		ReasoningTokens:  u.ReasoningTokens(),
		Cost:             u.Cost,
	}
}

func deriveTitle(text string, attachments []store.Attachment) string {
	title := strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	if title == "" && len(attachments) > 0 {
		title = attachments[0].Filename
	}
	runes := []rune(title)
	if len(runes) > 60 {
		title = strings.TrimSpace(string(runes[:60])) + "…"
	}
	return title
}

func humanizeStreamError(err error) string {
	var apiErr *openrouter.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Error()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "OpenRouter не ответил за отведённое время"
	}
	return "ошибка обращения к OpenRouter: " + err.Error()
}
