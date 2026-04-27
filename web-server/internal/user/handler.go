package user

import (
	"net/http"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/organization"
	"github.com/ke1ta1to/wallet-note/internal/shared/httpx"
)

type Handler struct {
	memRepo organization.MembershipRepository
	orgRepo organization.OrgRepository
	mw      *auth.Middleware
}

func New(memRepo organization.MembershipRepository, orgRepo organization.OrgRepository, mw *auth.Middleware) *Handler {
	return &Handler{memRepo: memRepo, orgRepo: orgRepo, mw: mw}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /me", h.mw.WithAuth(h.GetMe))
	mux.HandleFunc("GET /me/orgs", h.mw.WithAuth(h.GetMyOrgs))
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	var username *string
	if claims.Username != "" {
		username = &claims.Username
	}
	httpx.WriteJSON(w, http.StatusOK, MeResponse{
		UserID:   claims.Sub,
		Username: username,
	})
}

func (h *Handler) GetMyOrgs(w http.ResponseWriter, r *http.Request) {
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
	orgs, err := h.orgRepo.BatchGetOrgs(r.Context(), orgIDs)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	orgByID := make(map[string]string, len(orgs))
	for _, o := range orgs {
		orgByID[o.OrgID] = o.Name
	}
	for _, m := range memberships {
		items = append(items, organization.ToMembershipResponse(m, orgByID[m.OrgID]))
	}
	httpx.WriteJSON(w, http.StatusOK, MyOrgsResponse{Items: items})
}
