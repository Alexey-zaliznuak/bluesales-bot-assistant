package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("запись ответа", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Error: message})
}

// writeStoreError переводит ошибки хранилища в HTTP-коды.
func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "не найдено")
	default:
		slog.Error("ошибка хранилища", "error", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 4<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
		return false
	}
	return true
}
