// Package repository defines how the webhook module talks to storage,
// independent of which database engine backs it.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"webhook-middleware/internal/modules/webhook/domain"
)

// Filter describes the query options the dashboard list endpoint supports.
type Filter struct {
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
func (f *Filter) Normalize() {
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

// SourceCount is one row of the "distinct sources" breakdown.
type SourceCount struct {
	Source string `json:"source"`
	Total  int64  `json:"total"`
}

// StatusCount is one row of the "by publish status" breakdown.
type StatusCount struct {
	PublishStatus string `json:"publish_status"`
	Total         int64  `json:"total"`
}

// DailyCount is one row of the "events per day" time series.
type DailyCount struct {
	Day   string `json:"day"`
	Total int64  `json:"total"`
}

// Stats aggregates the numbers the dashboard overview page needs.
type Stats struct {
	TotalEvents   int64          `json:"total_events"`
	BySource      []SourceCount  `json:"by_source"`
	ByStatus      []StatusCount  `json:"by_status"`
	ByDay         []DailyCount   `json:"by_day"`
	FailedPending int64          `json:"failed_pending"`
}

// WebhookLogRepository persists and queries webhook callback records.
type WebhookLogRepository interface {
	Create(ctx context.Context, log *domain.WebhookLog) error
	UpdatePublishResult(ctx context.Context, id uuid.UUID, status domain.PublishStatus, publishErr string, topics []string) error
	IncrementRetry(ctx context.Context, id uuid.UUID) error

	FindByID(ctx context.Context, id uuid.UUID) (*domain.WebhookLog, error)
	Find(ctx context.Context, filter Filter) ([]domain.WebhookLog, int64, error)
	FindForRetry(ctx context.Context, statuses []domain.PublishStatus, maxRetry, limit int) ([]domain.WebhookLog, error)

	Sources(ctx context.Context) ([]SourceCount, error)
	Stats(ctx context.Context, from, to time.Time) (Stats, error)
}
