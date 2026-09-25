package repository

import (
	"context"
	"errors"
	"time"

	"webhook-middleware/internal/modules/webhook/domain"
	"webhook-middleware/pkg/shared"
	shareddomain "webhook-middleware/pkg/shared/domain"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/tracer"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type webhookLogRepoSQL struct {
	readDB, writeDB *gorm.DB
}

// NewWebhookLogRepoSQL webhook log repo constructor
func NewWebhookLogRepoSQL(readDB, writeDB *gorm.DB) WebhookLogRepository {
	return &webhookLogRepoSQL{readDB: readDB, writeDB: writeDB}
}

func (r *webhookLogRepoSQL) write(ctx context.Context) *gorm.DB {
	db := r.writeDB
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*gorm.DB); ok {
		db = tx
	}
	return shared.SetSpanToGorm(ctx, db)
}

func (r *webhookLogRepoSQL) Create(ctx context.Context, log *shareddomain.WebhookLog) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookLogRepoSQL:Create")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	return r.write(ctx).Create(log).Error
}

func (r *webhookLogRepoSQL) UpdatePublishResult(ctx context.Context, id uuid.UUID, status shareddomain.PublishStatus, publishErr string, topics []string) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookLogRepoSQL:UpdatePublishResult")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	updates := map[string]any{
		"publish_status": status,
		"publish_error":  publishErr,
		"kafka_topics":   shareddomain.StringArray(topics),
	}
	if status == shareddomain.PublishStatusPublished {
		updates["published_at"] = time.Now()
	}
	return r.write(ctx).
		Model(&shareddomain.WebhookLog{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *webhookLogRepoSQL) IncrementRetry(ctx context.Context, id uuid.UUID) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookLogRepoSQL:IncrementRetry")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	return r.write(ctx).
		Model(&shareddomain.WebhookLog{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

// FindByID reads from the primary: it is used for read-after-write flows
// (replay, bulk retry) where a lagging replica would return a stale status.
func (r *webhookLogRepoSQL) FindByID(ctx context.Context, id uuid.UUID) (result *shareddomain.WebhookLog, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookLogRepoSQL:FindByID")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	var log shareddomain.WebhookLog
	if err := shared.SetSpanToGorm(ctx, r.writeDB).First(&log, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &log, nil
}

func (r *webhookLogRepoSQL) applyFilter(q *gorm.DB, f *domain.FilterWebhookLog) *gorm.DB {
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

func (r *webhookLogRepoSQL) Find(ctx context.Context, f *domain.FilterWebhookLog) (logs []shareddomain.WebhookLog, total int64, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookLogRepoSQL:Find")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	f.Normalize()

	db := shared.SetSpanToGorm(ctx, r.readDB)
	if err = r.applyFilter(db.Model(&shareddomain.WebhookLog{}), f).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := f.SortBy
	if f.SortDesc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	err = r.applyFilter(db.Model(&shareddomain.WebhookLog{}), f).
		Order(order).
		Limit(f.Limit).
		Offset(f.CalculateOffset()).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *webhookLogRepoSQL) FindForRetry(ctx context.Context, statuses []shareddomain.PublishStatus, maxRetry, limit int) (logs []shareddomain.WebhookLog, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookLogRepoSQL:FindForRetry")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	err = shared.SetSpanToGorm(ctx, r.writeDB).
		Where("publish_status IN ?", statuses).
		Where("retry_count < ?", maxRetry).
		Order("created_at ASC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *webhookLogRepoSQL) Sources(ctx context.Context) (out []domain.SourceCount, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookLogRepoSQL:Sources")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	err = shared.SetSpanToGorm(ctx, r.readDB).
		Model(&shareddomain.WebhookLog{}).
		Select("source, COUNT(*) as total").
		Group("source").
		Order("total DESC").
		Scan(&out).Error
	return out, err
}

func (r *webhookLogRepoSQL) Stats(ctx context.Context, from, to time.Time) (stats domain.Stats, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "WebhookLogRepoSQL:Stats")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	db := shared.SetSpanToGorm(ctx, r.readDB)
	base := db.Model(&shareddomain.WebhookLog{}).
		Where("created_at BETWEEN ? AND ?", from, to)

	if err = base.Session(&gorm.Session{}).Count(&stats.TotalEvents).Error; err != nil {
		return stats, err
	}

	if err = base.Session(&gorm.Session{}).
		Select("source, COUNT(*) as total").
		Group("source").
		Order("total DESC").
		Scan(&stats.BySource).Error; err != nil {
		return stats, err
	}

	if err = base.Session(&gorm.Session{}).
		Select("publish_status, COUNT(*) as total").
		Group("publish_status").
		Scan(&stats.ByStatus).Error; err != nil {
		return stats, err
	}

	if err = base.Session(&gorm.Session{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, COUNT(*) as total").
		Group("day").
		Order("day ASC").
		Scan(&stats.ByDay).Error; err != nil {
		return stats, err
	}

	if err = db.Model(&shareddomain.WebhookLog{}).
		Where("publish_status = ?", shareddomain.PublishStatusFailed).
		Count(&stats.FailedPending).Error; err != nil {
		return stats, err
	}

	if stats.BySource == nil {
		stats.BySource = []domain.SourceCount{}
	}
	if stats.ByStatus == nil {
		stats.ByStatus = []domain.StatusCount{}
	}
	if stats.ByDay == nil {
		stats.ByDay = []domain.DailyCount{}
	}
	return stats, nil
}
