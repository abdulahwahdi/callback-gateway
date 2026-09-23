// Package worker runs background jobs for the webhook module. Today that's
// a single ticker that retries publishing any webhook stuck in "failed"
// state, so a temporary Kafka outage doesn't lose events -- they were
// already durably saved to the database on ingest.
package worker

import (
	"context"
	"time"

	"webhook-middleware/internal/modules/webhook/usecase"
	"webhook-middleware/internal/pkg/logger"
)

// RetryWorker periodically retries webhook events that failed to publish.
type RetryWorker struct {
	uc       usecase.WebhookUsecase
	interval time.Duration
	batch    int
}

// NewRetryWorker builds a RetryWorker. interval is how often to sweep for
// failed events; batch is the max number retried per sweep.
func NewRetryWorker(uc usecase.WebhookUsecase, interval time.Duration, batch int) *RetryWorker {
	if interval <= 0 {
		interval = time.Minute
	}
	if batch <= 0 {
		batch = 50
	}
	return &RetryWorker{uc: uc, interval: interval, batch: batch}
}

// Run blocks, sweeping on a ticker until ctx is cancelled.
func (w *RetryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	logger.Info("webhook retry worker started", "interval", w.interval.String(), "batch", w.batch)

	for {
		select {
		case <-ctx.Done():
			logger.Info("webhook retry worker stopped")
			return
		case <-ticker.C:
			w.sweep(ctx)
		}
	}
}

func (w *RetryWorker) sweep(ctx context.Context) {
	retried, failed, err := w.uc.RetryFailed(ctx, w.batch)
	if err != nil {
		logger.Error("retry worker sweep failed", "error", err)
		return
	}
	if retried > 0 || failed > 0 {
		logger.Info("retry worker sweep completed", "retried", retried, "still_failed", failed)
	}
}
