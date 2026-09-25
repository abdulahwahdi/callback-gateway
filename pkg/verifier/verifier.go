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
	"os"
	"strings"

	"github.com/golangid/candi/logger"
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

// NewRegistryFromEnv registers optional per-source verifiers from
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
func NewRegistryFromEnv() *Registry {
	registry := NewRegistry()

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

		upper := strings.ToUpper(source)
		mode := strings.ToLower(os.Getenv(prefix + upper + modeSuffix))
		header := os.Getenv(prefix + upper + "_HEADER")
		secret := os.Getenv(prefix + upper + "_SECRET")
		if header == "" || secret == "" {
			logger.LogYellow("incomplete webhook verifier config, skipping source: " + source)
			continue
		}

		switch mode {
		case "token":
			registry.Register(source, TokenVerifier{Header: header, Token: secret})
			logger.LogIf("registered token verifier: source=%s header=%s", source, header)
		case "hmac":
			registry.Register(source, HMACSHA256Verifier{Header: header, Secret: secret})
			logger.LogIf("registered hmac verifier: source=%s header=%s", source, header)
		default:
			logger.LogYellow("unknown webhook verifier mode, skipping source: " + source + " mode: " + mode)
		}
	}

	return registry
}
