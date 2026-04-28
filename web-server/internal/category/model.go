package category

import "time"

const (
	KindIncome  = "income"
	KindExpense = "expense"
)

type Category struct {
	ID        string
	OrgID     string
	Name      string
	Kind      string
	Color     string
	CreatedAt time.Time
}
