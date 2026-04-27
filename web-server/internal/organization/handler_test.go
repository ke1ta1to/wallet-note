package organization_test

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

	"github.com/ke1ta1to/wallet-note/internal/organization"
	"github.com/ke1ta1to/wallet-note/internal/testutils"
)

func TestPOSTOrgs_Success(t *testing.T) {
	mux, db := testutils.Setup(t)

	req := testutils.MakeReq(t, "POST", "/orgs",
		strings.NewReader(`{"name":"My Wallet"}`),
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var resp organization.OrgResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.OrgID == "" {
		t.Error("org_id empty")
	}
	if resp.CreatedAt == "" {
		t.Error("created_at empty")
	}
	want := organization.OrgResponse{Name: "My Wallet"}
	if diff := cmp.Diff(want, resp,
		cmpopts.IgnoreFields(organization.OrgResponse{}, "OrgID", "CreatedAt"),
	); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}

	// DDB: ORG meta + Membership both written.
	ctx := context.Background()
	orgItem, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(testutils.TableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: "ORG#" + resp.OrgID},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil || orgItem.Item == nil {
		t.Fatalf("ORG meta missing: err=%v", err)
	}
	memItem, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(testutils.TableName),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: "USER#user-1"},
			"SK": &ddbtypes.AttributeValueMemberS{Value: "ORG#" + resp.OrgID},
		},
	})
	if err != nil || memItem.Item == nil {
		t.Fatalf("Membership missing: err=%v", err)
	}
}

func TestPOSTOrgs_NoAuth(t *testing.T) {
	mux, _ := testutils.Setup(t)

	req := testutils.MakeReq(t, "POST", "/orgs",
		strings.NewReader(`{"name":"X"}`), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status=%d want 401", rec.Code)
	}
}

func TestPOSTOrgs_EmptyName(t *testing.T) {
	mux, _ := testutils.Setup(t)

	req := testutils.MakeReq(t, "POST", "/orgs",
		strings.NewReader(`{"name":""}`),
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d want 400", rec.Code)
	}
}

func TestGETOrg_AsMember(t *testing.T) {
	mux, _ := testutils.Setup(t)

	// Seed an org as user-1.
	createReq := testutils.MakeReq(t, "POST", "/orgs",
		strings.NewReader(`{"name":"Mine"}`),
		map[string]any{"sub": "user-1"},
	)
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("seed status=%d", createRec.Code)
	}
	var created organization.OrgResponse
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	// GET as the same user.
	req := testutils.MakeReq(t, "GET", "/orgs/"+created.OrgID, nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp organization.OrgResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(created, resp); diff != "" {
		t.Errorf("GET response should match POST response (-created +resp):\n%s", diff)
	}
}

func TestGETOrg_NotMember(t *testing.T) {
	mux, _ := testutils.Setup(t)

	// user-1 creates org.
	createReq := testutils.MakeReq(t, "POST", "/orgs",
		strings.NewReader(`{"name":"Theirs"}`),
		map[string]any{"sub": "user-1"},
	)
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	var created organization.OrgResponse
	json.NewDecoder(createRec.Body).Decode(&created)

	// user-2 tries to GET.
	req := testutils.MakeReq(t, "GET", "/orgs/"+created.OrgID, nil,
		map[string]any{"sub": "user-2"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status=%d want 403", rec.Code)
	}
}
