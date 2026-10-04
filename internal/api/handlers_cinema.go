package api

import (
	"net/http"

	"github.com/kazalab/zaphub/internal/model"
)

func (s *Server) handleCinemaReleases(w http.ResponseWriter, r *http.Request) {
	releases := s.cinemaStore.Releases()
	if releases == nil {
		releases = []model.CinemaRelease{}
	}
	writeJSON(w, http.StatusOK, releases)
}
