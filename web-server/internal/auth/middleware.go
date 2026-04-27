package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ke1ta1to/wallet-note/internal/platform/apperror"
	"github.com/ke1ta1to/wallet-note/internal/platform/httpx"
)

const (
	RoleOwner  = "owner"
	RoleMember = "member"
)

type Membership struct {
	UserID   string
	OrgID    string
	Role     string
	JoinedAt time.Time
}

// MembershipReader stays in auth so the package does not import organization.
type MembershipReader interface {
	Get(ctx context.Context, userID, orgID string) (*Membership, error)
}

type Middleware struct {
	reader MembershipReader
}

func NewMiddleware(reader MembershipReader) *Middleware {
	return &Middleware{reader: reader}
}

type ctxKey int

const (
	claimsCtxKey ctxKey = iota
	membershipCtxKey
)

// ClaimsFromContext is non-nil when the request passed through WithAuth (or
// WithOrgAuth/WithOwnerAuth that wrap it).
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsCtxKey).(*Claims)
	return c
}

// MembershipFromContext is non-nil when the request passed through
// WithOrgAuth (or WithOwnerAuth that wraps it).
func MembershipFromContext(ctx context.Context) *Membership {
	m, _ := ctx.Value(membershipCtxKey).(*Membership)
	return m
}

func (m *Middleware) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := ExtractClaims(r)
		if err != nil {
			httpx.WriteError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// WithOrgAuth requires {org_id} in the route pattern.
func (m *Middleware) WithOrgAuth(next http.HandlerFunc) http.HandlerFunc {
	return m.WithAuth(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		orgID := r.PathValue("org_id")
		if orgID == "" {
			httpx.WriteError(w, apperror.ErrInvalidInput)
			return
		}
		membership, err := m.reader.Get(r.Context(), claims.Sub, orgID)
		if errors.Is(err, apperror.ErrNotFound) {
			// Map "not a member" to 403; 404 would leak org existence.
			httpx.WriteError(w, apperror.ErrForbidden)
			return
		}
		if err != nil {
			httpx.WriteError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), membershipCtxKey, membership)
		next(w, r.WithContext(ctx))
	})
}

func (m *Middleware) WithOwnerAuth(next http.HandlerFunc) http.HandlerFunc {
	return m.WithOrgAuth(func(w http.ResponseWriter, r *http.Request) {
		mem := MembershipFromContext(r.Context())
		if mem.Role != RoleOwner {
			httpx.WriteError(w, apperror.ErrForbidden)
			return
		}
		next(w, r)
	})
}
