package organization

import (
	"context"

	"github.com/ke1ta1to/wallet-note/internal/auth"
)

type OrgRepository interface {
	GetOrg(ctx context.Context, orgID string) (*Organization, error)
	BatchGetOrgs(ctx context.Context, orgIDs []string) ([]*Organization, error)
}

// MembershipRepository implicitly satisfies auth.MembershipReader.
type MembershipRepository interface {
	GetMembership(ctx context.Context, userID, orgID string) (*auth.Membership, error)
	ListByUser(ctx context.Context, userID string) ([]*auth.Membership, error)
}
