package resthandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"webhook-middleware/pkg/shared"
)

func TestDashboardAuth(t *testing.T) {
	old := shared.GetEnv()
	t.Cleanup(func() { shared.SetEnv(old) })
	shared.SetEnv(shared.Environment{DashboardUsername: "admin", DashboardPassword: "pw"})

	h := dashboardAuthMiddleware()(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) { rw.WriteHeader(http.StatusOK) }))
	do := func(auth string) int {
		req := httptest.NewRequest(http.MethodGet, "/dashboard/webhooks", nil)
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	tok, _ := issueToken("admin", time.Hour)
	expired, _ := issueToken("admin", -time.Hour)
	parts := strings.Split(tok, ".")
	tampered := "9999999999." + parts[1] + "." + parts[2]

	if got := do(""); got != http.StatusUnauthorized {
		t.Errorf("no token: %d", got)
	}
	if got := do("Bearer " + tok); got != http.StatusOK {
		t.Errorf("valid token: %d", got)
	}
	if got := do("Bearer " + expired); got != http.StatusUnauthorized {
		t.Errorf("expired token: %d", got)
	}
	if got := do("Bearer " + tampered); got != http.StatusUnauthorized {
		t.Errorf("tampered token: %d", got)
	}
}
