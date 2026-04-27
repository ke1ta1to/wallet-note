package user_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/organization"
	"github.com/ke1ta1to/wallet-note/internal/testutils"
	"github.com/ke1ta1to/wallet-note/internal/user"
)

func TestGETMe(t *testing.T) {
	mux, _ := testutils.Setup(t)

	req := testutils.MakeReq(t, "GET", "/me", nil,
		map[string]any{"sub": "user-1", "username": "alice"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var got map[string]any
	json.NewDecoder(rec.Body).Decode(&got)
	want := map[string]any{
		"id":       "user-1",
		"username": "alice",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}
}

func TestGETMe_NoAuth(t *testing.T) {
	mux, _ := testutils.Setup(t)

	req := testutils.MakeReq(t, "GET", "/me", nil, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status=%d want 401", rec.Code)
	}
}

func TestGETMyOrgs_Empty(t *testing.T) {
	mux, _ := testutils.Setup(t)

	req := testutils.MakeReq(t, "GET", "/me/orgs", nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var resp user.MyOrgsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	want := user.MyOrgsResponse{Items: []organization.MembershipResponse{}}
	if diff := cmp.Diff(want, resp); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}
}

func TestGETMyOrgs_WithOrgs(t *testing.T) {
	mux, _ := testutils.Setup(t)

	// Create two orgs as user-1.
	for _, name := range []string{"First", "Second"} {
		req := testutils.MakeReq(t, "POST", "/orgs",
			strings.NewReader(`{"name":"`+name+`"}`),
			map[string]any{"sub": "user-1"},
		)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("seed %q status=%d", name, rec.Code)
		}
	}

	req := testutils.MakeReq(t, "GET", "/me/orgs", nil,
		map[string]any{"sub": "user-1"},
	)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp user.MyOrgsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	// Sort by OrgName so the comparison is deterministic.
	sort.Slice(resp.Items, func(i, j int) bool {
		return *resp.Items[i].OrgName < *resp.Items[j].OrgName
	})

	first, second := "First", "Second"
	want := []organization.MembershipResponse{
		{OrgName: &first, Role: auth.RoleOwner},
		{OrgName: &second, Role: auth.RoleOwner},
	}
	if diff := cmp.Diff(want, resp.Items,
		cmpopts.IgnoreFields(organization.MembershipResponse{}, "OrgID", "JoinedAt"),
	); diff != "" {
		t.Errorf("items mismatch (-want +got):\n%s", diff)
	}
}
