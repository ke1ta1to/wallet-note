package transaction

import "time"

type Transaction struct {
	ID         string
	OrgID      string
	CategoryID string
	Amount     int64
	Date       string // YYYY-MM-DD
	Memo       string
	CreatedAt  time.Time
	CreatedBy  string
}
