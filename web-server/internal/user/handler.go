package user

import (
	"net/http"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/shared/httpx"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /me", h.GetMe)
}

type MeResponse struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ExtractClaims(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, MeResponse{
		UserID:   claims.Sub,
		Username: claims.Username,
	})
}
