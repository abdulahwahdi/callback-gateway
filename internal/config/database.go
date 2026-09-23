package config

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"webhook-middleware/internal/modules/webhook/domain"
)

// NewDatabase opens the Postgres connection and ensures the webhook_logs
// table (and its indexes) exist via GORM AutoMigrate. For production use
// with a dedicated migration tool, see migrations/ -- the SQL there matches
// this schema.
func NewDatabase(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	if err := db.AutoMigrate(&domain.WebhookLog{}); err != nil {
		return nil, fmt.Errorf("auto-migrate webhook_logs: %w", err)
	}

	return db, nil
}
