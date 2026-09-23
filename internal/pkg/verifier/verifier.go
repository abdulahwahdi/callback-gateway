// Package verifier provides an optional, per-source signature/token check
// for inbound payment-gateway callbacks. It is intentionally pluggable:
// most gateways use either a static token header (e.g. Xendit's
// "x-callback-token") or an HMAC signature over the raw body (e.g. Stripe,
// Coinbase-style gateways). A source with no verifier configured is treated
// as "not checked" (nil), not "invalid" — ingestion never blocks on this.
package verifier

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

// Result of a signature check.
type Result struct {
	Checked  bool
	Verified bool
	Header   string
}

// Verifier validates one source's callback authenticity.
type Verifier interface {
	Verify(headers http.Header, body []byte) Result
}

// Registry maps a source name (as used in POST /webhooks/:source) to its verifier.
type Registry struct {
	verifiers map[string]Verifier
}

// NewRegistry builds an empty registry.
func NewRegistry() *Registry {
	return &Registry{verifiers: make(map[string]Verifier)}
}

// Register attaches a verifier for a given source.
func (r *Registry) Register(source string, v Verifier) {
	r.verifiers[source] = v
}

// Verify checks a callback for the given source. If no verifier is
// registered for that source, it returns Result{Checked: false}.
func (r *Registry) Verify(source string, headers http.Header, body []byte) Result {
	v, ok := r.verifiers[source]
	if !ok {
		return Result{Checked: false}
	}
	return v.Verify(headers, body)
}

// TokenVerifier checks that a header equals a static shared secret
// (e.g. Xendit's "x-callback-token", Doku's shared secret headers).
type TokenVerifier struct {
	Header string
	Token  string
}

func (t TokenVerifier) Verify(headers http.Header, _ []byte) Result {
	got := headers.Get(t.Header)
	ok := subtle.ConstantTimeCompare([]byte(got), []byte(t.Token)) == 1
	return Result{Checked: true, Verified: ok, Header: t.Header}
}

// HMACSHA256Verifier checks that a header contains the hex-encoded
// HMAC-SHA256 of the raw request body, signed with a shared secret
// (the scheme used by Stripe-like and many custom gateway integrations).
type HMACSHA256Verifier struct {
	Header string
	Secret string
}

func (h HMACSHA256Verifier) Verify(headers http.Header, body []byte) Result {
	sig := headers.Get(h.Header)
	mac := hmac.New(sha256.New, []byte(h.Secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	ok := subtle.ConstantTimeCompare([]byte(sig), []byte(expected)) == 1
	return Result{Checked: true, Verified: ok, Header: h.Header}
}
