package cinema

import (
	"context"
	"errors"
	"testing"

	"github.com/kazalab/zaphub/internal/model"
)

type fakeProvider struct {
	name     string
	releases []model.CinemaRelease
	err      error
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) Fetch(_ context.Context) ([]model.CinemaRelease, error) {
	return f.releases, f.err
}

func TestRefresherSuccess(t *testing.T) {
	provider := &fakeProvider{
		name:     "test",
		releases: []model.CinemaRelease{{ID: "1", Title: "Film A"}, {ID: "2", Title: "Film B"}},
	}
	store := NewStore("test")
	r := NewRefresher(provider, store)

	r.Refresh(context.Background())

	releases := store.Releases()
	if len(releases) != 2 {
		t.Fatalf("expected 2 releases, got %d", len(releases))
	}
	if releases[0].Title != "Film A" {
		t.Errorf("expected Film A, got %q", releases[0].Title)
	}
	_, _, err, _ := store.Status()
	if err != "" {
		t.Errorf("expected no error, got %q", err)
	}
}

func TestRefresherError(t *testing.T) {
	provider := &fakeProvider{name: "test", err: errors.New("network timeout")}
	store := NewStore("test")
	r := NewRefresher(provider, store)

	r.Refresh(context.Background())

	releases := store.Releases()
	if len(releases) != 0 {
		t.Fatalf("expected 0 releases, got %d", len(releases))
	}
	_, _, err, _ := store.Status()
	if err != "network timeout" {
		t.Errorf("expected error message, got %q", err)
	}
}

func TestRefresherEmptyList(t *testing.T) {
	provider := &fakeProvider{name: "test", releases: []model.CinemaRelease{}}
	store := NewStore("test")
	r := NewRefresher(provider, store)

	r.Refresh(context.Background())

	releases := store.Releases()
	if len(releases) != 0 {
		t.Fatalf("expected 0 releases, got %d", len(releases))
	}
	_, _, err, _ := store.Status()
	if err != "empty release list" {
		t.Errorf("expected empty error, got %q", err)
	}
}

func TestNewProviderUnknown(t *testing.T) {
	_, err := NewProvider("bogus", "http://example.com", nil)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNewProviderAllocine(t *testing.T) {
	p, err := NewProvider("allocine", "http://example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "allocine" {
		t.Errorf("expected allocine, got %q", p.Name())
	}
}

func TestStoreStatus(t *testing.T) {
	store := NewStore("myprovider")
	lastRefresh, lastSuccess, lastErr, provider := store.Status()
	if provider != "myprovider" {
		t.Errorf("provider = %q", provider)
	}
	if !lastRefresh.IsZero() || !lastSuccess.IsZero() {
		t.Error("expected zero times initially")
	}
	if lastErr != "" {
		t.Errorf("expected empty error, got %q", lastErr)
	}
}
