package user

import "github.com/ke1ta1to/wallet-note/internal/organization"

type MeResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type MyOrgsResponse struct {
	Items []organization.MembershipResponse `json:"items"`
}
