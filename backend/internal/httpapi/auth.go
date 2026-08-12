package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alex/bluesales-bot-assistant/backend/internal/auth"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

func contextWithTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

func validateRegistration(login, password string) error {
	loginLength := utf8.RuneCountInString(login)
	if loginLength < 3 || loginLength > 64 {
		return errors.New("логин должен содержать от 3 до 64 символов")
	}
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("пароль должен содержать не менее 8 символов")
	}
	if len([]byte(password)) > 72 {
		return errors.New("пароль не должен превышать 72 байта")
	}
	return nil
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if err := validateRegistration(req.Login, req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		slog.Error("хеширование пароля", "error", err)
		writeError(w, http.StatusInternalServerError, "не удалось создать пользователя")
		return
	}

	user, err := s.store.CreateUser(r.Context(), req.Login, passwordHash)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeError(w, http.StatusConflict, "этот логин уже занят")
			return
		}
		writeStoreError(w, err)
		return
	}

	if !s.createSession(w, r, user) {
		return
	}
	writeJSON(w, http.StatusCreated, userResponse{ID: user.ID, Login: user.Login})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "укажите логин и пароль")
		return
	}

	user, err := s.store.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// Не разделяем «нет пользователя» и «неверный пароль».
			writeError(w, http.StatusUnauthorized, "неверный логин или пароль")
			return
		}
		writeStoreError(w, err)
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "неверный логин или пароль")
		return
	}

	if !s.createSession(w, r, user) {
		return
	}
	writeJSON(w, http.StatusOK, userResponse{ID: user.ID, Login: user.Login})
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request, user *store.User) bool {
	token, tokenHash, err := auth.NewSessionToken()
	if err != nil {
		slog.Error("генерация токена сессии", "error", err)
		writeError(w, http.StatusInternalServerError, "не удалось создать сессию")
		return false
	}

	expiresAt := time.Now().Add(s.cfg.SessionTTL)
	if err := s.store.CreateSession(r.Context(), user.ID, tokenHash, expiresAt); err != nil {
		writeStoreError(w, err)
		return false
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return true
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(s.cfg.SessionCookieName); err == nil && cookie.Value != "" {
		if err := s.store.DeleteSession(r.Context(), auth.HashToken(cookie.Value)); err != nil {
			slog.Error("удаление сессии", "error", err)
		}
	}
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	writeJSON(w, http.StatusOK, userResponse{ID: user.ID, Login: user.Login})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
