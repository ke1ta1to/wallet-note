package transaction_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ke1ta1to/wallet-note/internal/category"
	"github.com/ke1ta1to/wallet-note/internal/organization"
	"github.com/ke1ta1to/wallet-note/internal/testutils"
	"github.com/ke1ta1to/wallet-note/internal/transaction"
)

// seedOrg creates an org owned by the given user and returns its id.
func seedOrg(t *testing.T, mux http.Handler, userSub string) string {
	t.Helper()
	req := testutils.MakeReq(t, "POST", "/orgs",
		strings.NewReader(`{"name":"Seed"}`),
		map[string]any{"sub": userSub},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed org status=%d body=%s", rec.Code, rec.Body.String())
	}
	var org organization.OrgResponse
	if err := json.NewDecoder(rec.Body).Decode(&org); err != nil {
		t.Fatalf("decode org: %v", err)
	}
	return org.ID
}

// seedCategory creates a category in the given org and returns its id.
func seedCategory(t *testing.T, mux http.Handler, userSub, orgID, name, kind string) string {
	t.Helper()
	body := fmt.Sprintf(`{"name":%q,"kind":%q,"color":"#000000"}`, name, kind)
	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/categories",
		strings.NewReader(body),
		map[string]any{"sub": userSub},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed cat status=%d body=%s", rec.Code, rec.Body.String())
	}
	var c category.CategoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatalf("decode cat: %v", err)
	}
	return c.ID
}

func TestPOSTTx_Success(t *testing.T) {
	mux, db := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")
	catID := seedCategory(t, mux, "user-1", orgID, "食費", "expense")

	body := fmt.Sprintf(`{"category_id":%q,"amount":1500,"date":"2026-04-25","memo":"昼"}`, catID)
	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/transactions",
		strings.NewReader(body),
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp transaction.TransactionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" || resp.CreatedAt == "" {
		t.Errorf("missing id/created_at: %+v", resp)
	}
	want := transaction.TransactionResponse{
		CategoryID: catID,
		Amount:     1500,
		Date:       "2026-04-25",
		Memo:       "昼",
	}
	if diff := cmp.Diff(want, resp,
		cmpopts.IgnoreFields(transaction.TransactionResponse{}, "ID", "CreatedAt"),
	); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}

	// DDB: tx item written under ORG#<orgID>/TX#<date>#<id>.
	ctx := context.Background()
	out, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(testutils.TableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: "ORG#" + orgID},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "TX#2026-04-25#" + resp.ID},
		},
	})
	if err != nil || out.Item == nil {
		t.Fatalf("tx item missing: err=%v", err)
	}
}

func TestPOSTTx_NoAuth(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/transactions",
		strings.NewReader(`{"category_id":"x","amount":100,"date":"2026-04-25"}`),
		nil,
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status=%d want 401", rec.Code)
	}
}

func TestPOSTTx_NotMember(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/transactions",
		strings.NewReader(`{"category_id":"x","amount":100,"date":"2026-04-25"}`),
		map[string]any{"sub": "user-2"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status=%d want 403", rec.Code)
	}
}

func TestPOSTTx_InvalidDate(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/transactions",
		strings.NewReader(`{"category_id":"x","amount":100,"date":"2026/04/25"}`),
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d want 400", rec.Code)
	}
}

func TestPOSTTx_NonPositiveAmount(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/transactions",
		strings.NewReader(`{"category_id":"x","amount":0,"date":"2026-04-25"}`),
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d want 400", rec.Code)
	}
}

func TestGETTxs_NoMonth(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "GET", "/orgs/"+orgID+"/transactions", nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d want 400", rec.Code)
	}
}

func TestGETTxs_NotMember(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "GET", "/orgs/"+orgID+"/transactions?month=2026-04", nil,
		map[string]any{"sub": "user-2"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status=%d want 403", rec.Code)
	}
}

func TestGETTxs_Empty(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "GET", "/orgs/"+orgID+"/transactions?month=2026-04", nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp transaction.TransactionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected empty, got %d items", len(resp.Items))
	}
}

func TestGETTxs_ByMonth(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")
	catFood := seedCategory(t, mux, "user-1", orgID, "食費", "expense")
	catSalary := seedCategory(t, mux, "user-1", orgID, "給与", "income")

	// Two txs in 2026-04, one in 2026-03 (must NOT appear in 2026-04 list).
	for _, body := range []string{
		fmt.Sprintf(`{"category_id":%q,"amount":1500,"date":"2026-04-10","memo":""}`, catFood),
		fmt.Sprintf(`{"category_id":%q,"amount":300000,"date":"2026-04-25","memo":""}`, catSalary),
		fmt.Sprintf(`{"category_id":%q,"amount":900,"date":"2026-03-30","memo":""}`, catFood),
	} {
		r := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/transactions",
			strings.NewReader(body),
			map[string]any{"sub": "user-1"},
		)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("seed tx status=%d body=%s", w.Code, w.Body.String())
		}
	}

	req := testutils.MakeReq(t, "GET", "/orgs/"+orgID+"/transactions?month=2026-04", nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp transaction.TransactionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	// SK ascending = date ascending.
	if resp.Items[0].Date != "2026-04-10" || resp.Items[1].Date != "2026-04-25" {
		t.Errorf("unexpected order/dates: %+v", resp.Items)
	}
}

func TestGETTxs_ByCategory(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")
	catFood := seedCategory(t, mux, "user-1", orgID, "食費", "expense")
	catSalary := seedCategory(t, mux, "user-1", orgID, "給与", "income")

	for _, body := range []string{
		fmt.Sprintf(`{"category_id":%q,"amount":1500,"date":"2026-04-10","memo":""}`, catFood),
		fmt.Sprintf(`{"category_id":%q,"amount":900,"date":"2026-04-15","memo":""}`, catFood),
		fmt.Sprintf(`{"category_id":%q,"amount":300000,"date":"2026-04-25","memo":""}`, catSalary),
	} {
		r := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/transactions",
			strings.NewReader(body),
			map[string]any{"sub": "user-1"},
		)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("seed tx status=%d body=%s", w.Code, w.Body.String())
		}
	}

	url := "/orgs/" + orgID + "/transactions?month=2026-04&category_id=" + catFood
	req := testutils.MakeReq(t, "GET", url, nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp transaction.TransactionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 food items, got %d", len(resp.Items))
	}
	for _, item := range resp.Items {
		if item.CategoryID != catFood {
			t.Errorf("unexpected category in result: %+v", item)
		}
	}
}
