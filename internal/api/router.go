package api

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"time"

	"github.com/kazalab/zaphub/internal/cinema"
	"github.com/kazalab/zaphub/internal/epg"
	"github.com/kazalab/zaphub/internal/youtube"
)

// Options configures the API server.
type Options struct {
	EPGStore        *epg.Store
	CinemaStore     *cinema.Store
	YTStore         *youtube.Store
	YTRefresh       func(ctx context.Context, ch youtube.Channel)
	WebFS           fs.FS
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	epgStore        *epg.Store
	cinemaStore     *cinema.Store
	ytStore         *youtube.Store
	ytRefresh       func(ctx context.Context, ch youtube.Channel)
	mux             *http.ServeMux
	shutdownTimeout time.Duration
	httpServer      *http.Server
}

// New builds the API server. ytRefresh is called after a channel is added via
// the API to resolve its name and fetch its latest videos; it may be nil.
// opts.WebFS serves static assets at "/" (the embedded web UI); it may be nil
// to disable the web UI.
func New(opts Options) *Server {
	s := &Server{
		epgStore:        opts.EPGStore,
		cinemaStore:     opts.CinemaStore,
		ytStore:         opts.YTStore,
		ytRefresh:       opts.YTRefresh,
		shutdownTimeout: opts.ShutdownTimeout,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/epg/channels", s.handleEPGChannels)
	mux.HandleFunc("GET /api/epg/now", s.handleEPGNow)
	mux.HandleFunc("GET /api/epg/evening", s.handleEPGEvening)
	mux.HandleFunc("GET /api/epg/programmes", s.handleEPGProgrammes)
	mux.HandleFunc("GET /api/cinema/releases", s.handleCinemaReleases)
	mux.HandleFunc("GET /api/youtube/channels", s.handleYTChannels)
	mux.HandleFunc("POST /api/youtube/channels", s.handleYTAddChannel)
	mux.HandleFunc("DELETE /api/youtube/channels/{id}", s.handleYTDeleteChannel)
	mux.HandleFunc("GET /api/youtube/videos", s.handleYTVideos)
	if opts.WebFS != nil {
		mux.Handle("GET /", http.FileServer(http.FS(opts.WebFS)))
	}
	s.mux = mux

	s.httpServer = &http.Server{
		Addr:         opts.Addr,
		Handler:      s.mux,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		IdleTimeout:  opts.IdleTimeout,
	}

	return s
}

// Run starts the HTTP server and blocks until ctx is cancelled, then shuts
// down gracefully with the configured shutdown timeout. A shutdown via ctx
// returns nil.
func (s *Server) Run(ctx context.Context) error {
	//nolint:gosec // G118 - shutdown goroutine intentionally uses context.Background for graceful server shutdown.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
	}()
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Handler() http.Handler {
	return s.mux
}
