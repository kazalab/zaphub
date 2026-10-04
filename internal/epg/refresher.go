package epg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var errEmpty = errors.New("empty channel list")

type Refresher struct {
	fetcher *Fetcher
	store   *Store
}

func NewRefresher(f *Fetcher, s *Store) *Refresher {
	return &Refresher{fetcher: f, store: s}
}

// Refresh fetches the EPG and updates the store. It keeps the current cache
// when upstream did not change and rejects empty channel lists.
func (r *Refresher) Refresh(ctx context.Context) error {
	channels, programmes, changed, err := r.fetcher.Fetch(ctx)
	if err != nil {
		r.store.SetError(err)
		return fmt.Errorf("epg: refresh failed: %w", err)
	}
	if !changed {
		slog.Info("epg: not modified, keeping cache")
		return nil
	}
	if len(channels) == 0 {
		r.store.SetError(errEmpty)
		return errEmpty
	}
	r.store.Update(channels, programmes)
	slog.Info("epg: refreshed", "channels", len(channels), "programmes", len(programmes))
	return nil
}
