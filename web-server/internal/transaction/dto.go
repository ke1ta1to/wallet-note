package transaction

import "time"

type CreateTransactionRequest struct {
	CategoryID string `json:"category_id" validate:"required"`
	Amount     int64  `json:"amount" validate:"required,min=1"`
	Date       string `json:"date" validate:"required,datetime=2006-01-02"`
	Memo       string `json:"memo" validate:"max=500"`
}

type TransactionResponse struct {
	ID         string `json:"id"`
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount"`
	Date       string `json:"date"`
	Memo       string `json:"memo"`
	CreatedAt  string `json:"created_at"`
}

type TransactionsResponse struct {
	Items      []TransactionResponse `json:"items"`
	NextCursor string                `json:"nextCursor,omitempty"`
}

func ToTransactionResponse(t *Transaction) TransactionResponse {
	return TransactionResponse{
		ID:         t.ID,
		CategoryID: t.CategoryID,
		Amount:     t.Amount,
		Date:       t.Date,
		Memo:       t.Memo,
		CreatedAt:  t.CreatedAt.UTC().Format(time.RFC3339),
	}
}
