package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/ke1ta1to/wallet-note/internal/platform/apperror"
)

var validate = validator.New()

func DecodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return apperror.ErrInvalidInput
	}
	if err := validate.Struct(dst); err != nil {
		return apperror.ErrInvalidInput
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write json", "err", err)
	}
}

func WriteError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, apperror.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, apperror.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, apperror.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, apperror.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, apperror.ErrInvalidInput):
		status = http.StatusBadRequest
	}
	if status == http.StatusInternalServerError {
		slog.Error("internal error", "err", err)
	}
	WriteJSON(w, status, errorBody{Message: err.Error()})
}

type errorBody struct {
	Message string `json:"message"`
}
