package category

import (
	"context"
	"net/http"
	"time"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/platform/httpx"
	"github.com/ke1ta1to/wallet-note/internal/platform/idgen"
)

type categoryRepo interface {
	Put(ctx context.Context, c *Category) error
	ListByOrg(ctx context.Context, orgID string) ([]*Category, error)
}

type Handler struct {
	repo categoryRepo
	mw   *auth.Middleware
}

func New(repo categoryRepo, mw *auth.Middleware) *Handler {
	return &Handler{repo: repo, mw: mw}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /orgs/{org_id}/categories", h.mw.WithOrgAuth(h.Create))
	mux.HandleFunc("GET /orgs/{org_id}/categories", h.mw.WithOrgAuth(h.List))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	mem := auth.MembershipFromContext(r.Context())
	var req CreateCategoryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	c := &Category{
		ID:        idgen.NewID(),
		OrgID:     mem.OrgID,
		Name:      req.Name,
		Kind:      req.Kind,
		Color:     req.Color,
		CreatedAt: time.Now().UTC(),
	}
	if err := h.repo.Put(r.Context(), c); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, ToCategoryResponse(c))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	mem := auth.MembershipFromContext(r.Context())
	cats, err := h.repo.ListByOrg(r.Context(), mem.OrgID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items := make([]CategoryResponse, 0, len(cats))
	for _, c := range cats {
		items = append(items, ToCategoryResponse(c))
	}
	httpx.WriteJSON(w, http.StatusOK, CategoriesResponse{Items: items})
}
