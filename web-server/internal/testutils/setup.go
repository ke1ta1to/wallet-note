package testutils

import (
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/ke1ta1to/wallet-note/internal/app"
)

// Setup wires the application identically to main.go and resets the table.
func Setup(t *testing.T) (*http.ServeMux, *dynamodb.Client) {
	t.Helper()
	db := DDB(t)
	ResetTable(t, db)
	mux := app.NewMux(db, TableName)
	return mux, db
}
