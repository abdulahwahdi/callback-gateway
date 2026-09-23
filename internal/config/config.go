// Package config loads and validates this service's runtime configuration
// from environment variables (see .env.sample for every supported key).
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"webhook-middleware/internal/pkg/logger"
	"webhook-middleware/internal/pkg/verifier"
)

// Config holds every environment-derived setting the service needs.
type Config struct {
	HTTPPort string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
	DBSSLMode string

	KafkaBrokers     []string
	KafkaCommonTopic string
	KafkaTopicPrefix string
	MaxPublishRetries int

	RetryWorkerInterval time.Duration
	RetryWorkerBatch    int

	// DashboardAPIKey, if set, is required (via X-API-Key header) to call
	// any /dashboard/* endpoint. Left empty in local/dev by default.
	DashboardAPIKey string

	Verifiers *verifier.Registry
}

// Load reads .env (if present, ignored if missing) then environment
// variables, applying defaults suitable for local development.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		logger.Info("no .env file found, relying on process environment")
	}

	cfg := &Config{
		HTTPPort: getEnv("HTTP_PORT", "8090"),

		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASS", "postgres"),
		DBName:    getEnv("DB_NAME", "webhook_middleware"),
		DBSSLMode: getEnv("DB_SSLMODE", "disable"),

		KafkaBrokers:      strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaCommonTopic:  getEnv("KAFKA_COMMON_TOPIC", "webhook_gateway_events"),
		KafkaTopicPrefix:  getEnv("KAFKA_TOPIC_PREFIX", "webhook_gateway"),
		MaxPublishRetries: getEnvInt("MAX_PUBLISH_RETRIES", 5),

		RetryWorkerInterval: getEnvDuration("RETRY_WORKER_INTERVAL", time.Minute),
		RetryWorkerBatch:    getEnvInt("RETRY_WORKER_BATCH", 50),

		DashboardAPIKey: os.Getenv("DASHBOARD_API_KEY"),
	}

	cfg.Verifiers = buildVerifiers()

	return cfg
}

// buildVerifiers registers optional per-source signature verifiers from
// environment variables:
//
//	WEBHOOK_VERIFY_<SOURCE>_MODE=token|hmac
//	WEBHOOK_VERIFY_<SOURCE>_HEADER=<header-name>
//	WEBHOOK_VERIFY_<SOURCE>_SECRET=<shared-secret-or-token>
//
// e.g. for Xendit's static callback token:
//
//	WEBHOOK_VERIFY_XENDIT_MODE=token
//	WEBHOOK_VERIFY_XENDIT_HEADER=x-callback-token
//	WEBHOOK_VERIFY_XENDIT_SECRET=your-xendit-callback-token
//
// A source with no WEBHOOK_VERIFY_<SOURCE>_* set is simply not checked.
func buildVerifiers() *verifier.Registry {
	registry := verifier.NewRegistry()

	const prefix = "WEBHOOK_VERIFY_"
	const modeSuffix = "_MODE"

	seen := map[string]bool{}
	for _, kv := range os.Environ() {
		key, _, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, modeSuffix) {
			continue
		}
		source := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(key, prefix), modeSuffix))
		if source == "" || seen[source] {
			continue
		}
		seen[source] = true

		mode := strings.ToLower(getEnv(prefix+strings.ToUpper(source)+modeSuffix, ""))
		header := getEnv(prefix+strings.ToUpper(source)+"_HEADER", "")
		secret := getEnv(prefix+strings.ToUpper(source)+"_SECRET", "")
		if header == "" || secret == "" {
			logger.Warn("incomplete webhook verifier config, skipping", "source", source)
			continue
		}

		switch mode {
		case "token":
			registry.Register(source, verifier.TokenVerifier{Header: header, Token: secret})
			logger.Info("registered token verifier", "source", source, "header", header)
		case "hmac":
			registry.Register(source, verifier.HMACSHA256Verifier{Header: header, Secret: secret})
			logger.Info("registered hmac verifier", "source", source, "header", header)
		default:
			logger.Warn("unknown verifier mode, skipping", "source", source, "mode", mode)
		}
	}

	return registry
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
