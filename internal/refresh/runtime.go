package refresh

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/kazalab/zaphub/internal/cinema"
	"github.com/kazalab/zaphub/internal/epg"
	"github.com/kazalab/zaphub/internal/youtube"
)

// Options configures the refresh runtime.
type Options struct {
	EPGInterval    time.Duration
	CinemaInterval time.Duration
	YTInterval     time.Duration
	EPG            *epg.Refresher
	Cinema         *cinema.Refresher
	YT             *youtube.Refresher
}

// Runtime keeps the data stores fresh. A boot refresh of every source makes
// the API serve from cache as soon as possible; the periodic refreshes keep
// the stores up to date.
type Runtime struct {
	opts Options
}

func New(opts Options) *Runtime {
	return &Runtime{opts: opts}
}

// Run refreshes all data sources at boot (an EPG failure aborts the startup),
// then starts the periodic refreshes until ctx is done. Periodic refresh
// failures are logged and do not stop the runtime.
func (r *Runtime) Run(ctx context.Context) error {
	if err := r.opts.EPG.Refresh(ctx); err != nil {
		return err
	}
	r.opts.Cinema.Refresh(ctx)
	r.opts.YT.Refresh(ctx)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return runTicker(ctx, r.opts.EPGInterval, func() {
			if err := r.opts.EPG.Refresh(ctx); err != nil {
				slog.Error("epg: refresh failed", "error", err)
			}
		})
	})
	g.Go(func() error {
		return runTicker(ctx, r.opts.CinemaInterval, func() {
			r.opts.Cinema.Refresh(ctx)
		})
	})
	g.Go(func() error {
		return runTicker(ctx, r.opts.YTInterval, func() {
			r.opts.YT.Refresh(ctx)
		})
	})

	return g.Wait()
}

// runTicker calls fn periodically until ctx is cancelled.
func runTicker(ctx context.Context, interval time.Duration, fn func()) error {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			fn()
		}
	}
}
