// Package logger provides a tiny structured logging wrapper around log/slog
// so the rest of the service doesn't depend directly on the stdlib logging
// package (keeps it easy to swap for zap/zerolog/candi's logger later).
package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func init() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	log = slog.New(handler)
}

// L returns the shared structured logger instance.
func L() *slog.Logger { return log }

func Info(msg string, args ...any)  { log.Info(msg, args...) }
func Warn(msg string, args ...any)  { log.Warn(msg, args...) }
func Error(msg string, args ...any) { log.Error(msg, args...) }
func Debug(msg string, args ...any) { log.Debug(msg, args...) }
