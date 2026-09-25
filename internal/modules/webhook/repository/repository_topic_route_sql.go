package repository

import (
	"context"
	"errors"
	"fmt"

	"webhook-middleware/internal/modules/webhook/domain"
	"webhook-middleware/pkg/shared"
	shareddomain "webhook-middleware/pkg/shared/domain"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/tracer"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type topicRouteRepoSQL struct {
	readDB, writeDB *gorm.DB
}

// NewTopicRouteRepoSQL topic route repo constructor
func NewTopicRouteRepoSQL(readDB, writeDB *gorm.DB) TopicRouteRepository {
	return &topicRouteRepoSQL{readDB: readDB, writeDB: writeDB}
}

func (r *topicRouteRepoSQL) write(ctx context.Context) *gorm.DB {
	db := r.writeDB
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*gorm.DB); ok {
		db = tx
	}
	return shared.SetSpanToGorm(ctx, db)
}

func (r *topicRouteRepoSQL) Create(ctx context.Context, route *shareddomain.TopicRoute) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TopicRouteRepoSQL:Create")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	if err := r.write(ctx).Create(route).Error; err != nil {
		return fmt.Errorf("create topic route: %w", err)
	}
	return nil
}

func (r *topicRouteRepoSQL) Update(ctx context.Context, id uuid.UUID, in *domain.RequestUpdateTopicRoute) (route *shareddomain.TopicRoute, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TopicRouteRepoSQL:Update")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	updates := map[string]any{}
	if in.Env != nil {
		updates["env"] = *in.Env
	}
	if in.Source != nil {
		updates["source"] = *in.Source
	}
	if in.Topic != nil {
		updates["topic"] = *in.Topic
	}
	if in.Enabled != nil {
		updates["enabled"] = *in.Enabled
	}

	if len(updates) > 0 {
		if err := r.write(ctx).
			Model(&shareddomain.TopicRoute{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update topic route: %w", err)
		}
	}

	return r.FindByID(ctx, id)
}

func (r *topicRouteRepoSQL) Delete(ctx context.Context, id uuid.UUID) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TopicRouteRepoSQL:Delete")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	return r.write(ctx).Delete(&shareddomain.TopicRoute{}, "id = ?", id).Error
}

// FindByID reads from the primary so it can follow a write (Update) safely.
func (r *topicRouteRepoSQL) FindByID(ctx context.Context, id uuid.UUID) (route *shareddomain.TopicRoute, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TopicRouteRepoSQL:FindByID")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	var found shareddomain.TopicRoute
	if err := shared.SetSpanToGorm(ctx, r.writeDB).First(&found, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &found, nil
}

func (r *topicRouteRepoSQL) List(ctx context.Context) (routes []shareddomain.TopicRoute, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TopicRouteRepoSQL:List")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	err = shared.SetSpanToGorm(ctx, r.readDB).Order("env ASC, source ASC").Find(&routes).Error
	return routes, err
}

func (r *topicRouteRepoSQL) ListEnabled(ctx context.Context) (routes []shareddomain.TopicRoute, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TopicRouteRepoSQL:ListEnabled")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	err = shared.SetSpanToGorm(ctx, r.readDB).Where("enabled = ?", true).Find(&routes).Error
	return routes, err
}
