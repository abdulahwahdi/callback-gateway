package domain

import (
	"time"

	"github.com/google/uuid"
)

// TopicRoute is a DB-configurable override for which Kafka topic a given
// (environment, source) pair publishes to, layered on top of the
// env-var-driven template (see usecase.TopicResolver). This is what lets an
// operator point e.g. "production" + "midtrans" at a different topic than
// "staging" + "midtrans" -- or any source -- without redeploying.
//
// Env and Source may each be "*" as a wildcard:
//
//	("*", "midtrans")     -- applies to "midtrans" in every environment
//	("production", "*")   -- applies to every source in "production"
//
// Resolution always prefers the most specific match: exact (env, source) >
// (env, "*") > ("*", source) > the env-var template.
type TopicRoute struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Env    string    `json:"env" gorm:"type:varchar(50);uniqueIndex:idx_topic_routes_env_source"`
	Source string    `json:"source" gorm:"type:varchar(100);uniqueIndex:idx_topic_routes_env_source"`
	Topic  string    `json:"topic" gorm:"type:varchar(255)"`
	// Enabled lets an override be disabled (fall back to the template)
	// without deleting the row.
	Enabled bool `json:"enabled" gorm:"default:true"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName pins the GORM table name explicitly.
func (TopicRoute) TableName() string { return "kafka_topic_routes" }

// NewTopicRoute builds an enabled topic route ready to be persisted.
func NewTopicRoute(env, source, topic string) *TopicRoute {
	return &TopicRoute{
		ID:      uuid.New(),
		Env:     env,
		Source:  source,
		Topic:   topic,
		Enabled: true,
	}
}
