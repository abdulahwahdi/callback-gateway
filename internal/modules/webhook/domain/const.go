package domain

import "errors"

// Wildcard matches "any env" / "any source" in a DB topic route override.
const Wildcard = "*"

// ErrNotFound is returned by repositories when a record does not exist.
var ErrNotFound = errors.New("record not found")

// ErrInvalidTopicRoute is returned when env, source, or topic is empty.
var ErrInvalidTopicRoute = errors.New("env, source, and topic are all required")
