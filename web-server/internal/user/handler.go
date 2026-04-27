package user

import (
	"context"
	"net/http"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/organization"
	"github.com/ke1ta1to/wallet-note/internal/platform/httpx"
)

type membershipReader interface {
	ListByUser(ctx context.Context, userID string) ([]*auth.Membership, error)
}

type orgBatchReader interface {
	BatchGet(ctx context.Context, orgIDs []string) ([]*organization.Organization, error)
}

type Handler struct {
	memRepo membershipReader
	orgRepo orgBatchReader
	mw      *auth.Middleware
}

func New(memRepo membershipReader, orgRepo orgBatchReader, mw *auth.Middleware) *Handler {
	return &Handler{memRepo: memRepo, orgRepo: orgRepo, mw: mw}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /me", h.mw.WithAuth(h.Me))
	mux.HandleFunc("GET /me/orgs", h.mw.WithAuth(h.MyOrgs))
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	var username *string
	if claims.Username != "" {
		username = &claims.Username
	}
	httpx.WriteJSON(w, http.StatusOK, MeResponse{
		ID:       claims.Sub,
		Username: username,
	})
}

func (h *Handler) MyOrgs(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	memberships, err := h.memRepo.ListByUser(r.Context(), claims.Sub)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	items := make([]organization.MembershipResponse, 0, len(memberships))
	if len(memberships) == 0 {
		httpx.WriteJSON(w, http.StatusOK, MyOrgsResponse{Items: items})
		return
	}

	orgIDs := make([]string, 0, len(memberships))
	for _, m := range memberships {
		orgIDs = append(orgIDs, m.OrgID)
	}
	orgs, err := h.orgRepo.BatchGet(r.Context(), orgIDs)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	orgByID := make(map[string]string, len(orgs))
	for _, o := range orgs {
		orgByID[o.ID] = o.Name
	}
	for _, m := range memberships {
		items = append(items, organization.ToMembershipResponse(m, orgByID[m.OrgID]))
	}
	httpx.WriteJSON(w, http.StatusOK, MyOrgsResponse{Items: items})
}
