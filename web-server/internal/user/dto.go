package user

import "github.com/ke1ta1to/wallet-note/internal/organization"

type MeResponse struct {
	UserID   string  `json:"user_id"`
	Username *string `json:"username,omitempty"`
}

type MyOrgsResponse struct {
	Items []organization.MembershipResponse `json:"items"`
}
