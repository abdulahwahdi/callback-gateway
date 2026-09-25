package resthandler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"webhook-middleware/pkg/shared"

	"github.com/golangid/candi/wrapper"
)

var (
	randomSecretOnce sync.Once
	randomSecret     []byte
)

// signingSecret returns the HMAC key for session tokens: DASHBOARD_AUTH_SECRET
// when set, otherwise a per-process random key.
func signingSecret() []byte {
	if s := shared.GetEnv().DashboardAuthSecret; s != "" {
		return []byte(s)
	}
	randomSecretOnce.Do(func() {
		randomSecret = make([]byte, 32)
		_, _ = rand.Read(randomSecret)
	})
	return randomSecret
}

func loginEnabled() bool {
	e := shared.GetEnv()
	return e.DashboardUsername != "" && e.DashboardPassword != ""
}

func sign(payload string) string {
	m := hmac.New(sha256.New, signingSecret())
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// issueToken returns "<expiry-unix>.<username>.<signature>".
func issueToken(username string, ttl time.Duration) (string, time.Time) {
	exp := time.Now().Add(ttl)
	payload := strconv.FormatInt(exp.Unix(), 10) + "." + base64.RawURLEncoding.EncodeToString([]byte(username))
	return payload + "." + sign(payload), exp
}

func validToken(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	payload := parts[0] + "." + parts[1]
	if subtle.ConstantTimeCompare([]byte(sign(payload)), []byte(parts[2])) != 1 {
		return false
	}
	exp, err := strconv.ParseInt(parts[0], 10, 64)
	return err == nil && time.Now().Unix() < exp
}

func secureEqual(a, b string) bool {
	ha, hb := sha256.Sum256([]byte(a)), sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(ha[:], hb[:]) == 1
}

// dashboardAuthMiddleware guards /dashboard/* routes. Access is granted by a
// valid bearer token (from /auth/login) or a matching X-API-Key. If neither
// DASHBOARD_API_KEY nor DASHBOARD_USERNAME/PASSWORD is configured, the
// dashboard stays open (local development).
func dashboardAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			apiKey := shared.GetEnv().DashboardAPIKey
			if apiKey == "" && !loginEnabled() {
				next.ServeHTTP(rw, req)
				return
			}
			if apiKey != "" && secureEqual(req.Header.Get("X-API-Key"), apiKey) {
				next.ServeHTTP(rw, req)
				return
			}
			if tok, ok := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer "); ok && loginEnabled() && validToken(tok) {
				next.ServeHTTP(rw, req)
				return
			}
			wrapper.NewHTTPResponse(http.StatusUnauthorized, "authentication required").JSON(rw)
		})
	}
}

// authStatus handles GET /auth/status -- tells the frontend whether to show the login page.
func (h *RestHandler) authStatus(rw http.ResponseWriter, req *http.Request) {
	wrapper.NewHTTPResponse(http.StatusOK, "ok", map[string]any{
		"login_enabled": loginEnabled(),
		"auth_required": loginEnabled() || shared.GetEnv().DashboardAPIKey != "",
	}).JSON(rw)
}

// login handles POST /auth/login with {"username","password"}.
func (h *RestHandler) login(rw http.ResponseWriter, req *http.Request) {
	if !loginEnabled() {
		wrapper.NewHTTPResponse(http.StatusNotFound, "login is not enabled").JSON(rw)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(rw, req.Body, 4096)).Decode(&body); err != nil {
		wrapper.NewHTTPResponse(http.StatusBadRequest, "invalid request body").JSON(rw)
		return
	}
	e := shared.GetEnv()
	// Evaluate both so timing doesn't reveal which field was wrong.
	userOK := secureEqual(body.Username, e.DashboardUsername)
	passOK := secureEqual(body.Password, e.DashboardPassword)
	if !userOK || !passOK {
		wrapper.NewHTTPResponse(http.StatusUnauthorized, "invalid username or password").JSON(rw)
		return
	}
	token, exp := issueToken(body.Username, e.DashboardSessionTTL)
	wrapper.NewHTTPResponse(http.StatusOK, "ok", map[string]any{
		"token":      token,
		"expires_at": exp.UTC().Format(time.RFC3339),
		"username":   body.Username,
	}).JSON(rw)
}
