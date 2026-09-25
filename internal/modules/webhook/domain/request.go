package domain

import "net/http"

// IngestRequest is everything the REST layer extracts from an inbound HTTP
// callback before handing it to the usecase.
type IngestRequest struct {
	Env      string // optional; empty means the service default (APP_ENV)
	Source   string
	Method   string
	Endpoint string
	SourceIP string
	Headers  http.Header
	Query    map[string][]string
	Body     []byte
}

// RequestCreateTopicRoute is the payload for adding a topic route override.
type RequestCreateTopicRoute struct {
	Env    string `json:"env"`
	Source string `json:"source"`
	Topic  string `json:"topic"`
}

// RequestUpdateTopicRoute carries only the fields the caller wants to change;
// nil fields are left untouched.
type RequestUpdateTopicRoute struct {
	Env     *string `json:"env"`
	Source  *string `json:"source"`
	Topic   *string `json:"topic"`
	Enabled *bool   `json:"enabled"`
}
