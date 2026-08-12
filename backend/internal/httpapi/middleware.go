package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/alex/bluesales-bot-assistant/backend/internal/auth"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

type ctxKey string

const userCtxKey ctxKey = "user"

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(s.cfg.SessionCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "нужна авторизация")
			return
		}

		user, err := s.store.GetUserBySessionToken(r.Context(), auth.HashToken(cookie.Value))
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				s.clearSessionCookie(w)
				writeError(w, http.StatusUnauthorized, "сессия истекла")
				return
			}
			writeStoreError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userFrom(ctx context.Context) *store.User {
	user, _ := ctx.Value(userCtxKey).(*store.User)
	return user
}

func (s *Server) isAdmin(user *store.User) bool {
	return user != nil &&
		strings.TrimSpace(s.cfg.SeedUserLogin) != "" &&
		strings.EqualFold(strings.TrimSpace(user.Login), strings.TrimSpace(s.cfg.SeedUserLogin))
}

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.isAdmin(userFrom(r.Context())) {
			writeError(w, http.StatusForbidden, "доступ разрешён только администратору")
			return
		}
		next.ServeHTTP(w, r)
	})
}
