package organization

import (
	"time"

	"github.com/ke1ta1to/wallet-note/internal/auth"
)

type CreateOrgRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type OrgResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func ToOrgResponse(o *Organization) OrgResponse {
	return OrgResponse{
		ID:        o.ID,
		Name:      o.Name,
		CreatedAt: o.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type MembershipResponse struct {
	OrgID    string  `json:"org_id"`
	OrgName  *string `json:"org_name,omitempty"`
	Role     string  `json:"role"`
	JoinedAt string  `json:"joined_at"`
}

func ToMembershipResponse(m *auth.Membership, orgName string) MembershipResponse {
	var name *string
	if orgName != "" {
		name = &orgName
	}
	return MembershipResponse{
		OrgID:    m.OrgID,
		OrgName:  name,
		Role:     m.Role,
		JoinedAt: m.JoinedAt.UTC().Format(time.RFC3339),
	}
}
