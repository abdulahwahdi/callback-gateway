package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"webhook-middleware/internal/modules/webhook/domain"
)

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository builds a WebhookLogRepository backed by PostgreSQL/GORM.
func NewPostgresRepository(db *gorm.DB) WebhookLogRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, log *domain.WebhookLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *postgresRepository) UpdatePublishResult(ctx context.Context, id uuid.UUID, status domain.PublishStatus, publishErr string, topics []string) error {
	updates := map[string]any{
		"publish_status": status,
		"publish_error":  publishErr,
		"kafka_topics":   StringArrayValue(topics),
	}
	if status == domain.PublishStatusPublished {
		now := time.Now()
		updates["published_at"] = now
	}
	return r.db.WithContext(ctx).
		Model(&domain.WebhookLog{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *postgresRepository) IncrementRetry(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.WebhookLog{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.WebhookLog, error) {
	var log domain.WebhookLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *postgresRepository) applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.Source != "" {
		q = q.Where("source = ?", f.Source)
	}
	if f.PublishStatus != "" {
		q = q.Where("publish_status = ?", f.PublishStatus)
	}
	if f.EventType != "" {
		q = q.Where("event_type = ?", f.EventType)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("source ILIKE ? OR event_type ILIKE ? OR endpoint ILIKE ?", like, like, like)
	}
	return q
}

func (r *postgresRepository) Find(ctx context.Context, f Filter) ([]domain.WebhookLog, int64, error) {
	f.Normalize()

	base := r.applyFilter(r.db.WithContext(ctx).Model(&domain.WebhookLog{}), f)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := f.SortBy
	if f.SortDesc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	var logs []domain.WebhookLog
	err := r.applyFilter(r.db.WithContext(ctx).Model(&domain.WebhookLog{}), f).
		Order(order).
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *postgresRepository) FindForRetry(ctx context.Context, statuses []domain.PublishStatus, maxRetry, limit int) ([]domain.WebhookLog, error) {
	var logs []domain.WebhookLog
	err := r.db.WithContext(ctx).
		Where("publish_status IN ?", statuses).
		Where("retry_count < ?", maxRetry).
		Order("created_at ASC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *postgresRepository) Sources(ctx context.Context) ([]SourceCount, error) {
	var out []SourceCount
	err := r.db.WithContext(ctx).
		Model(&domain.WebhookLog{}).
		Select("source, COUNT(*) as total").
		Group("source").
		Order("total DESC").
		Scan(&out).Error
	return out, err
}

func (r *postgresRepository) Stats(ctx context.Context, from, to time.Time) (Stats, error) {
	var stats Stats

	base := r.db.WithContext(ctx).Model(&domain.WebhookLog{}).
		Where("created_at BETWEEN ? AND ?", from, to)

	if err := base.Session(&gorm.Session{}).Count(&stats.TotalEvents).Error; err != nil {
		return stats, err
	}

	if err := base.Session(&gorm.Session{}).
		Select("source, COUNT(*) as total").
		Group("source").
		Order("total DESC").
		Scan(&stats.BySource).Error; err != nil {
		return stats, err
	}

	if err := base.Session(&gorm.Session{}).
		Select("publish_status, COUNT(*) as total").
		Group("publish_status").
		Scan(&stats.ByStatus).Error; err != nil {
		return stats, err
	}

	if err := base.Session(&gorm.Session{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, COUNT(*) as total").
		Group("day").
		Order("day ASC").
		Scan(&stats.ByDay).Error; err != nil {
		return stats, err
	}

	if err := r.db.WithContext(ctx).Model(&domain.WebhookLog{}).
		Where("publish_status = ?", domain.PublishStatusFailed).
		Count(&stats.FailedPending).Error; err != nil {
		return stats, err
	}

	if stats.BySource == nil {
		stats.BySource = []SourceCount{}
	}
	if stats.ByStatus == nil {
		stats.ByStatus = []StatusCount{}
	}
	if stats.ByDay == nil {
		stats.ByDay = []DailyCount{}
	}

	return stats, nil
}

// StringArrayValue is a tiny helper so callers can pass a plain []string
// into a jsonb update without importing the domain package's driver glue directly.
func StringArrayValue(topics []string) domain.StringArray {
	return domain.StringArray(topics)
}
