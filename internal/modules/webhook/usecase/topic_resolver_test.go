package usecase

import (
	"context"
	"testing"

	"webhook-middleware/internal/modules/webhook/domain"
	shareddomain "webhook-middleware/pkg/shared/domain"

	"github.com/google/uuid"
)

type fakeTopicRouteRepo struct {
	routes []shareddomain.TopicRoute
}

func (f *fakeTopicRouteRepo) Create(context.Context, *shareddomain.TopicRoute) error { return nil }
func (f *fakeTopicRouteRepo) Update(context.Context, uuid.UUID, *domain.RequestUpdateTopicRoute) (*shareddomain.TopicRoute, error) {
	return nil, nil
}
func (f *fakeTopicRouteRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (f *fakeTopicRouteRepo) FindByID(context.Context, uuid.UUID) (*shareddomain.TopicRoute, error) {
	return nil, nil
}
func (f *fakeTopicRouteRepo) List(context.Context) ([]shareddomain.TopicRoute, error) {
	return f.routes, nil
}
func (f *fakeTopicRouteRepo) ListEnabled(context.Context) ([]shareddomain.TopicRoute, error) {
	return f.routes, nil
}

func TestTopicResolver(t *testing.T) {
	repo := &fakeTopicRouteRepo{}
	r := NewTopicResolver(repo, "Production", "events_:env", "gw.:env.:source")

	if got := r.CommonTopic(""); got != "events_production" {
		t.Fatalf("CommonTopic = %q", got)
	}
	if got := r.SourceTopic("", "midtrans"); got != "gw.production.midtrans" {
		t.Fatalf("template fallback = %q", got)
	}

	repo.routes = []shareddomain.TopicRoute{
		{Env: "*", Source: "midtrans", Topic: "any-env-midtrans"},
		{Env: "production", Source: "*", Topic: "prod-all"},
	}
	if err := r.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	// (env, "*") beats ("*", source)
	if got := r.SourceTopic("", "midtrans"); got != "prod-all" {
		t.Fatalf("(env,*) should win over (*,source), got %q", got)
	}

	repo.routes = append(repo.routes, shareddomain.TopicRoute{Env: "production", Source: "midtrans", Topic: "exact"})
	_ = r.Refresh(context.Background())
	if got := r.SourceTopic("", "Midtrans"); got != "exact" {
		t.Fatalf("exact match should win, got %q", got)
	}
	if got := r.SourceTopic("", "xendit"); got != "prod-all" {
		t.Fatalf("(env,*) wildcard = %q", got)
	}

	// per-request env from the URL path overrides the service default
	repo.routes = []shareddomain.TopicRoute{{Env: "staging", Source: "dummy", Topic: "stg-dummy"}}
	_ = r.Refresh(context.Background())
	if got := r.SourceTopic("Staging", "dummy"); got != "stg-dummy" {
		t.Fatalf("path env route = %q", got)
	}
	if got := r.SourceTopic("staging", "other"); got != "gw.staging.other" {
		t.Fatalf("path env template = %q", got)
	}
	if got := r.CommonTopic("staging"); got != "events_staging" {
		t.Fatalf("path env common = %q", got)
	}
}
