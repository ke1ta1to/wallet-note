package transaction

import (
	"context"
	"net/http"
	"time"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/platform/apperror"
	"github.com/ke1ta1to/wallet-note/internal/platform/httpx"
	"github.com/ke1ta1to/wallet-note/internal/platform/idgen"
)

type txRepo interface {
	Put(ctx context.Context, t *Transaction) error
	ListByMonth(ctx context.Context, in ListByMonthInput) (*ListByMonthOutput, error)
}

type Handler struct {
	repo txRepo
	mw   *auth.Middleware
}

func New(repo txRepo, mw *auth.Middleware) *Handler {
	return &Handler{repo: repo, mw: mw}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /orgs/{org_id}/transactions", h.mw.WithOrgAuth(h.Create))
	mux.HandleFunc("GET /orgs/{org_id}/transactions", h.mw.WithOrgAuth(h.List))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	mem := auth.MembershipFromContext(r.Context())
	claims := auth.ClaimsFromContext(r.Context())
	var req CreateTransactionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	t := &Transaction{
		ID:         idgen.NewID(),
		OrgID:      mem.OrgID,
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		Date:       req.Date,
		Memo:       req.Memo,
		CreatedAt:  time.Now().UTC(),
		CreatedBy:  claims.Sub,
	}
	if err := h.repo.Put(r.Context(), t); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, ToTransactionResponse(t))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	mem := auth.MembershipFromContext(r.Context())
	q := r.URL.Query()
	month := q.Get("month")
	if _, err := time.Parse("2006-01", month); err != nil {
		httpx.WriteError(w, apperror.ErrInvalidInput)
		return
	}
	out, err := h.repo.ListByMonth(r.Context(), ListByMonthInput{
		OrgID:      mem.OrgID,
		Month:      month,
		CategoryID: q.Get("category_id"),
		Cursor:     q.Get("cursor"),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items := make([]TransactionResponse, 0, len(out.Items))
	for _, t := range out.Items {
		items = append(items, ToTransactionResponse(t))
	}
	httpx.WriteJSON(w, http.StatusOK, TransactionsResponse{Items: items, NextCursor: out.NextCursor})
}
