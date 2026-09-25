package domain

import "time"

// FilterWebhookLog describes the query options the dashboard list endpoint supports.
type FilterWebhookLog struct {
	Source        string
	PublishStatus string
	EventType     string
	From          *time.Time
	To            *time.Time
	Search        string // matches against source/event_type/endpoint
	Page          int
	Limit         int
	SortBy        string // "created_at" (default)
	SortDesc      bool
}

// Normalize fills in sane defaults/bounds for pagination.
func (f *FilterWebhookLog) Normalize() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 200 {
		f.Limit = 200
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
		f.SortDesc = true
	}
}

// CalculateOffset returns the row offset for the current page.
func (f *FilterWebhookLog) CalculateOffset() int {
	return (f.Page - 1) * f.Limit
}
