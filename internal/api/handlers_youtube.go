package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kazalab/zaphub/internal/model"
	"github.com/kazalab/zaphub/internal/youtube"
)

func (s *Server) handleYTChannels(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.ytStore.Channels())
}

func (s *Server) handleYTAddChannel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChannelID string `json:"channel_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.ChannelID = strings.TrimSpace(req.ChannelID)
	if req.ChannelID == "" {
		writeError(w, http.StatusBadRequest, "missing channel_id")
		return
	}

	added, err := s.ytStore.Add(youtube.Channel{ID: req.ChannelID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if s.ytRefresh != nil && added {
		go s.ytRefresh(context.Background(), youtube.Channel{ID: req.ChannelID})
	}

	ch, _ := s.ytStore.Get(req.ChannelID)
	status := http.StatusOK
	if added {
		status = http.StatusCreated
	}
	writeJSON(w, status, ch)
}

func (s *Server) handleYTDeleteChannel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	removed, err := s.ytStore.Remove(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !removed {
		writeError(w, http.StatusNotFound, "channel not found: "+id)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleYTVideos(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	var videos []model.Video
	if id := r.URL.Query().Get("channel"); id != "" {
		if _, ok := s.ytStore.Get(id); !ok {
			writeError(w, http.StatusNotFound, "channel not found: "+id)
			return
		}
		for _, v := range s.ytStore.Videos(id) {
			videos = append(videos, toModelVideo(v))
			if len(videos) >= limit {
				break
			}
		}
	} else {
		for _, v := range s.ytStore.AllVideos() {
			videos = append(videos, toModelVideo(v))
			if len(videos) >= limit {
				break
			}
		}
	}
	if videos == nil {
		videos = []model.Video{}
	}
	writeJSON(w, http.StatusOK, videos)
}

func toModelVideo(v youtube.Video) model.Video {
	return model.Video{
		ID:          v.ID,
		Title:       v.Title,
		Description: v.Description,
		Link:        v.Link,
		Published:   v.Published.Format(time.RFC3339),
		Updated:     v.Updated.Format(time.RFC3339),
		Thumbnail:   v.Thumbnail,
		Likes:       v.Likes,
		Duration:    v.Duration,
		ChannelID:   v.ChannelID,
		ChannelName: v.ChannelName,
	}
}
