package api

import (
	"net/http"
	"time"
)

func isoOrNull(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.Format(time.RFC3339)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	epgRefresh, epgSuccess, epgErr := s.epgStore.Status()
	ytRefresh, ytErr := s.ytStore.Status()
	cinemaRefresh, cinemaSuccess, cinemaErr, cinemaProvider := s.cinemaStore.Status()
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"epg": map[string]any{
			"last_refresh": isoOrNull(epgRefresh),
			"last_success": isoOrNull(epgSuccess),
			"last_error":   epgErr,
		},
		"cinema": map[string]any{
			"provider":     cinemaProvider,
			"releases":     len(s.cinemaStore.Releases()),
			"last_refresh": isoOrNull(cinemaRefresh),
			"last_success": isoOrNull(cinemaSuccess),
			"last_error":   cinemaErr,
		},
		"youtube": map[string]any{
			"channels":     len(s.ytStore.Channels()),
			"last_refresh": isoOrNull(ytRefresh),
			"last_error":   ytErr,
		},
	})
}
