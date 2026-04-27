package organization

import "time"

type Organization struct {
	OrgID     string
	Name      string
	CreatedAt time.Time
	CreatedBy string
}
