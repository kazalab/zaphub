package cinema

import (
	"context"
	"errors"
	"log/slog"
)

var errEmpty = errors.New("empty release list")

type Refresher struct {
	provider Provider
	store    *Store
}

func NewRefresher(provider Provider, store *Store) *Refresher {
	return &Refresher{provider: provider, store: store}
}

func (r *Refresher) Refresh(ctx context.Context) {
	releases, err := r.provider.Fetch(ctx)
	if err != nil {
		r.store.SetError(err)
		slog.Error("cinema: refresh failed", "provider", r.provider.Name(), "error", err)
		return
	}
	if len(releases) == 0 {
		r.store.SetError(errEmpty)
		slog.Warn("cinema: empty release list, keeping cache", "provider", r.provider.Name())
		return
	}
	r.store.Update(releases)
	slog.Info("cinema: refreshed", "provider", r.provider.Name(), "count", len(releases))
}
