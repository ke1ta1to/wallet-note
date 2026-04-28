package category_test

import (
	"context"
	"encoding/json"
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

func TestPOSTCategory_Success(t *testing.T) {
	mux, db := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/categories",
		strings.NewReader(`{"name":"食費","kind":"expense","color":"#ff8800"}`),
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp category.CategoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" || resp.CreatedAt == "" {
		t.Errorf("missing id/created_at: %+v", resp)
	}
	want := category.CategoryResponse{Name: "食費", Kind: "expense", Color: "#ff8800"}
	if diff := cmp.Diff(want, resp,
		cmpopts.IgnoreFields(category.CategoryResponse{}, "ID", "CreatedAt"),
	); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}

	// DDB: category item written under ORG#<orgID>/CATEGORY#<id>.
	ctx := context.Background()
	out, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(testutils.TableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: "ORG#" + orgID},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "CATEGORY#" + resp.ID},
		},
	})
	if err != nil || out.Item == nil {
		t.Fatalf("category item missing: err=%v", err)
	}
}

func TestPOSTCategory_NoAuth(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/categories",
		strings.NewReader(`{"name":"x","kind":"expense","color":"#000000"}`),
		nil,
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status=%d want 401", rec.Code)
	}
}

func TestPOSTCategory_NotMember(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/categories",
		strings.NewReader(`{"name":"x","kind":"expense","color":"#000000"}`),
		map[string]any{"sub": "user-2"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status=%d want 403", rec.Code)
	}
}

func TestPOSTCategory_InvalidKind(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/categories",
		strings.NewReader(`{"name":"x","kind":"savings","color":"#000000"}`),
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d want 400", rec.Code)
	}
}

func TestPOSTCategory_InvalidColor(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/categories",
		strings.NewReader(`{"name":"x","kind":"expense","color":"orange"}`),
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d want 400", rec.Code)
	}
}

func TestGETCategories_Empty(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "GET", "/orgs/"+orgID+"/categories", nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp category.CategoriesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected empty items, got %d", len(resp.Items))
	}
}

func TestGETCategories_AsMember(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	// Seed two categories.
	for _, body := range []string{
		`{"name":"食費","kind":"expense","color":"#ff8800"}`,
		`{"name":"給与","kind":"income","color":"#00aa55"}`,
	} {
		r := testutils.MakeReq(t, "POST", "/orgs/"+orgID+"/categories",
			strings.NewReader(body),
			map[string]any{"sub": "user-1"},
		)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("seed cat status=%d body=%s", w.Code, w.Body.String())
		}
	}

	req := testutils.MakeReq(t, "GET", "/orgs/"+orgID+"/categories", nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp category.CategoriesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	// UUIDv7 ascending → first inserted comes first.
	if resp.Items[0].Name != "食費" || resp.Items[1].Name != "給与" {
		t.Errorf("unexpected order: %+v", resp.Items)
	}
}

func TestGETCategories_NotMember(t *testing.T) {
	mux, _ := testutils.Setup(t)
	orgID := seedOrg(t, mux, "user-1")

	req := testutils.MakeReq(t, "GET", "/orgs/"+orgID+"/categories", nil,
		map[string]any{"sub": "user-2"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status=%d want 403", rec.Code)
	}
}
