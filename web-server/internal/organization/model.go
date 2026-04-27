package organization

import "time"

type Organization struct {
	ID        string
	Name      string
	CreatedAt time.Time
	CreatedBy string
}
