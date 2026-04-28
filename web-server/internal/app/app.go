// Package app exposes NewMux for both main.go and request tests so wiring
// stays consistent across them.
package app

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/category"
	"github.com/ke1ta1to/wallet-note/internal/organization"
	"github.com/ke1ta1to/wallet-note/internal/platform/router"
	"github.com/ke1ta1to/wallet-note/internal/transaction"
	"github.com/ke1ta1to/wallet-note/internal/user"
)

func NewMux(db *dynamodb.Client, tableName string) *http.ServeMux {
	orgRepo := organization.NewOrgRepo(db, tableName)
	memRepo := organization.NewMembershipRepo(db, tableName)
	orgSvc := organization.NewService(db, tableName)
	catRepo := category.NewCategoryRepo(db, tableName)
	txRepo := transaction.NewTxRepo(db, tableName)
	mw := auth.NewMiddleware(memRepo)

	return router.New(
		organization.New(orgRepo, orgSvc, mw),
		user.New(memRepo, orgRepo, mw),
		category.New(catRepo, mw),
		transaction.New(txRepo, mw),
	)
}
