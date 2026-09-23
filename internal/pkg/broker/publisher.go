// Package broker abstracts "publish this message to Kafka" so the usecase
// layer never imports segmentio/kafka-go directly (keeps it swappable/mockable).
package broker

import (
	"context"
	"sync"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// Publisher publishes a message to a topic, keyed for partitioning.
type Publisher interface {
	Publish(ctx context.Context, topic, key string, value []byte, headers map[string]string) error
	Close() error
}

type kafkaPublisher struct {
	brokers []string
	mu      sync.Mutex
	writers map[string]*kafka.Writer
}

// NewKafkaPublisher creates a Publisher that lazily opens one kafka.Writer
// per topic (kafka-go writers are topic-scoped and safe for concurrent use).
func NewKafkaPublisher(brokers []string) Publisher {
	return &kafkaPublisher{
		brokers: brokers,
		writers: make(map[string]*kafka.Writer),
	}
}

func (p *kafkaPublisher) writerFor(topic string) *kafka.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:                   kafka.TCP(p.brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		WriteTimeout:           5 * time.Second,
		BatchTimeout:           50 * time.Millisecond,
	}
	p.writers[topic] = w
	return w
}

func (p *kafkaPublisher) Publish(ctx context.Context, topic, key string, value []byte, headers map[string]string) error {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	msg := kafka.Message{
		Key:     []byte(key),
		Value:   value,
		Headers: kafkaHeaders,
		Time:    time.Now(),
	}

	return p.writerFor(topic).WriteMessages(ctx, msg)
}

func (p *kafkaPublisher) Close() error {
	var firstErr error
	for _, w := range p.writers {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
