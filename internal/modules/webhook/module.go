// Package webhook wires the module's layers together (domain, repository,
// usecase, delivery), following the same modular-monolith convention used
// across candi-based services: each business capability lives in its own
// self-contained module under internal/modules/<name>, exposing exactly the
// pieces main.go needs to mount it.
package webhook

import (
	"time"

	"gorm.io/gorm"

	"webhook-middleware/internal/modules/webhook/delivery/resthandler"
	"webhook-middleware/internal/modules/webhook/repository"
	"webhook-middleware/internal/modules/webhook/usecase"
	"webhook-middleware/internal/modules/webhook/worker"
	"webhook-middleware/internal/pkg/broker"
	"webhook-middleware/internal/pkg/verifier"
)

// Module bundles everything the webhook module needs to run: its REST
// handler (mounted onto the app's echo instance) and its background retry
// worker (run as a goroutine by main.go).
type Module struct {
	RestHandler *resthandler.RestHandler
	RetryWorker *worker.RetryWorker
	Usecase     usecase.WebhookUsecase
	publisher   broker.Publisher
}

// Close releases the module's resources (Kafka writers).
func (m *Module) Close() error {
	return m.publisher.Close()
}

// Options configures how the module's dependencies are built.
type Options struct {
	DB                 *gorm.DB
	KafkaBrokers        []string
	KafkaCommonTopic    string
	KafkaTopicPrefix    string
	MaxPublishRetries   int
	RetryWorkerInterval time.Duration
	RetryWorkerBatch    int
	Verifiers           *verifier.Registry
}

// New assembles the webhook module: repository -> usecase -> delivery/worker.
func New(opts Options) *Module {
	if opts.Verifiers == nil {
		opts.Verifiers = verifier.NewRegistry()
	}

	repo := repository.NewPostgresRepository(opts.DB)
	publisher := broker.NewKafkaPublisher(opts.KafkaBrokers)

	uc := usecase.New(repo, publisher, opts.Verifiers, usecase.KafkaConfig{
		CommonTopic: opts.KafkaCommonTopic,
		TopicPrefix: opts.KafkaTopicPrefix,
	}, opts.MaxPublishRetries)

	return &Module{
		RestHandler: resthandler.New(uc),
		RetryWorker: worker.NewRetryWorker(uc, opts.RetryWorkerInterval, opts.RetryWorkerBatch),
		Usecase:     uc,
		publisher:   publisher,
	}
}
