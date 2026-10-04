package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/kazalab/zaphub/internal/api"
	"github.com/kazalab/zaphub/internal/cinema"
	"github.com/kazalab/zaphub/internal/config"
	"github.com/kazalab/zaphub/internal/epg"
	"github.com/kazalab/zaphub/internal/refresh"
	"github.com/kazalab/zaphub/internal/youtube"
	"github.com/kazalab/zaphub/web"
)

func runDaemon(logLevel string, args []string) error {
	cfg := config.Load()

	if err := setupLogger(logLevel); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	slog.Info("starting zaphub api", "port", cfg.Port)

	client := &http.Client{Timeout: 60 * time.Second}

	epgStore := epg.NewStore()
	epgFetcher := epg.NewFetcher(cfg.EPGURL, client)
	epgRefresher := epg.NewRefresher(epgFetcher, epgStore)

	cinemaProvider, err := cinema.NewProvider(cfg.CinemaProvider, cfg.CinemaURL, client)
	if err != nil {
		return fmt.Errorf("cinema: provider: %w", err)
	}
	cinemaStore := cinema.NewStore(cfg.CinemaProvider)
	cinemaRefresher := cinema.NewRefresher(cinemaProvider, cinemaStore)

	ytStore := youtube.NewStore(cfg.YTChannelsFile)
	if err := ytStore.Load(); err != nil {
		slog.Warn("youtube: load channels", "error", err)
	}
	ytRefresher := youtube.NewRefresher(client, ytStore)

	rt := refresh.New(refresh.Options{
		EPGInterval:    cfg.EPGRefreshInterval,
		CinemaInterval: cfg.CinemaRefreshInterval,
		YTInterval:     cfg.YTRefreshInterval,
		EPG:            epgRefresher,
		Cinema:         cinemaRefresher,
		YT:             ytRefresher,
	})

	addr := ":" + cfg.Port
	srv := api.New(api.Options{
		EPGStore:        epgStore,
		CinemaStore:     cinemaStore,
		YTStore:         ytStore,
		YTRefresh:       ytRefresher.RefreshChannel,
		WebFS:           web.WebFS,
		Addr:            addr,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	})

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return rt.Run(ctx)
	})

	g.Go(func() error {
		slog.Info("zaphub listening", "addr", addr)
		return srv.Run(ctx)
	})

	return g.Wait()
}

func setupLogger(level string) error {
	var lvl slog.Level

	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	case "", "info":
		lvl = slog.LevelInfo
	default:
		return fmt.Errorf("invalid log level %q (debug, info, warn, error)", level)
	}

	slog.SetDefault(
		slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: lvl},
		)),
	)

	return nil
}
