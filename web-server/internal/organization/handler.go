package organization

import (
	"context"
	"net/http"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/platform/httpx"
)

type orgReader interface {
	Get(ctx context.Context, orgID string) (*Organization, error)
}

type Handler struct {
	repo orgReader
	svc  *Service
	mw   *auth.Middleware
}

func New(repo orgReader, svc *Service, mw *auth.Middleware) *Handler {
	return &Handler{repo: repo, svc: svc, mw: mw}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /orgs", h.mw.WithAuth(h.Create))
	mux.HandleFunc("GET /orgs/{orgId}", h.mw.WithOrgAuth(h.Get))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	var req CreateOrgRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	org, err := h.svc.CreateOrgWithMembership(r.Context(), claims.Sub, req.Name)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, ToOrgResponse(org))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	mem := auth.MembershipFromContext(r.Context())
	org, err := h.repo.Get(r.Context(), mem.OrgID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToOrgResponse(org))
}
