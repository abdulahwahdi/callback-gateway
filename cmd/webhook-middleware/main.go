// Command webhook-middleware runs the HTTP service that receives every
// payment-gateway webhook callback, durably logs it (source + raw body),
// fans it out to Kafka for every interested downstream service, and serves
// a dashboard API for tracking all of it.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"webhook-middleware/internal/config"
	webhookmodule "webhook-middleware/internal/modules/webhook"
	"webhook-middleware/internal/pkg/logger"
)

func main() {
	cfg := config.Load()

	db, err := config.NewDatabase(cfg)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	mod := webhookmodule.New(webhookmodule.Options{
		DB:                  db,
		KafkaBrokers:        cfg.KafkaBrokers,
		KafkaCommonTopic:    cfg.KafkaCommonTopic,
		KafkaTopicPrefix:    cfg.KafkaTopicPrefix,
		MaxPublishRetries:   cfg.MaxPublishRetries,
		RetryWorkerInterval: cfg.RetryWorkerInterval,
		RetryWorkerBatch:    cfg.RetryWorkerBatch,
		Verifiers:           cfg.Verifiers,
	})

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.BodyLimit("6M"))

	mod.RestHandler.Mount(e, config.DashboardAuth(cfg.DashboardAPIKey))

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	go mod.RetryWorker.Run(workerCtx)

	go func() {
		addr := ":" + cfg.HTTPPort
		logger.Info("webhook-middleware listening", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Error("http server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down webhook-middleware...")

	cancelWorker()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("error during http server shutdown", "error", err)
	}

	if err := mod.Close(); err != nil {
		logger.Error("error closing webhook module resources", "error", err)
	}

	logger.Info("shutdown complete")
}
