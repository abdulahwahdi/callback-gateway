package domain

import "github.com/google/uuid"

// ResponseIngest is what payment gateways get back after a callback is recorded.
type ResponseIngest struct {
	ID            uuid.UUID `json:"id"`
	Source        string    `json:"source"`
	PublishStatus string    `json:"publish_status"`
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
	TotalEvents   int64         `json:"total_events"`
	BySource      []SourceCount `json:"by_source"`
	ByStatus      []StatusCount `json:"by_status"`
	ByDay         []DailyCount  `json:"by_day"`
	FailedPending int64         `json:"failed_pending"`
}

// ResponseRetryFailed summarises a bulk retry run.
type ResponseRetryFailed struct {
	Retried int `json:"retried"`
	Failed  int `json:"failed"`
}

// Pagination describes pagination metadata attached to list endpoints.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalData  int64 `json:"total_data"`
	TotalPages int   `json:"total_pages"`
}

// NewPagination builds pagination metadata for the given page/limit/total.
func NewPagination(page, limit int, total int64) Pagination {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return Pagination{Page: page, Limit: limit, TotalData: total, TotalPages: totalPages}
}
