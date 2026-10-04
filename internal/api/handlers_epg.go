package api

import (
	"net/http"
	"sort"
	"time"
	_ "time/tzdata"

	"github.com/kazalab/zaphub/internal/epg"
	"github.com/kazalab/zaphub/internal/model"
)

var paris = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return time.FixedZone("CET", 3600)
	}
	return loc
}()

type nowEntry struct {
	ChannelID   string           `json:"channel_id"`
	ChannelName string           `json:"channel_name"`
	Icon        string           `json:"icon,omitempty"`
	Programme   *model.Programme `json:"programme"`
}

func (s *Server) handleEPGChannels(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.epgStore.Channels())
}

func (s *Server) handleEPGNow(w http.ResponseWriter, r *http.Request) {
	current := s.epgStore.Now(time.Now().In(paris))

	if id := r.URL.Query().Get("channel"); id != "" {
		ch, ok := s.epgStore.Channel(id)
		if !ok {
			writeError(w, http.StatusNotFound, "channel not found: "+id)
			return
		}
		writeJSON(w, http.StatusOK, buildNowEntry(ch, current[id]))
		return
	}

	entries := make([]nowEntry, 0, 30)
	for _, ch := range s.epgStore.Channels() {
		entries = append(entries, buildNowEntry(ch, current[ch.ID]))
	}
	writeJSON(w, http.StatusOK, entries)
}

func buildNowEntry(ch model.Channel, p epg.Programme) nowEntry {
	entry := nowEntry{ChannelID: ch.ID, ChannelName: ch.Name, Icon: ch.Icon}
	if p.Title != "" {
		prog := toModelProgramme(p)
		entry.Programme = &prog
	}
	return entry
}

func (s *Server) handleEPGEvening(w http.ResponseWriter, r *http.Request) {
	date := time.Now().In(paris)
	if d := r.URL.Query().Get("date"); d != "" {
		t, err := time.ParseInLocation("2006-01-02", d, paris)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date, expected YYYY-MM-DD")
			return
		}
		date = t
	}

	grouped := epg.EveningProgrammes(s.epgStore.Programmes(), date)
	resp := model.EveningResponse{Date: date.Format("2006-01-02")}

	if id := r.URL.Query().Get("channel"); id != "" {
		ch, ok := s.epgStore.Channel(id)
		if !ok {
			writeError(w, http.StatusNotFound, "channel not found: "+id)
			return
		}
		resp.Channels = []model.EveningChannel{buildEveningChannel(ch, grouped[id])}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	for _, ch := range s.epgStore.Channels() {
		resp.Channels = append(resp.Channels, buildEveningChannel(ch, grouped[ch.ID]))
	}
	writeJSON(w, http.StatusOK, resp)
}

func buildEveningChannel(ch model.Channel, progs []epg.Programme) model.EveningChannel {
	out := model.EveningChannel{ChannelID: ch.ID, ChannelName: ch.Name, Icon: ch.Icon}
	for _, p := range progs {
		out.Programmes = append(out.Programmes, toModelEveningProgramme(p))
	}
	return out
}

func (s *Server) handleEPGProgrammes(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("channel")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing required query parameter: channel")
		return
	}
	ch, ok := s.epgStore.Channel(id)
	if !ok {
		writeError(w, http.StatusNotFound, "channel not found: "+id)
		return
	}

	from := startOfDay(time.Now().In(paris))
	to := from.Add(24 * time.Hour)
	var err error
	if v := r.URL.Query().Get("from"); v != "" {
		if from, err = time.Parse(time.RFC3339, v); err != nil {
			writeError(w, http.StatusBadRequest, "invalid from, expected RFC3339")
			return
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if to, err = time.Parse(time.RFC3339, v); err != nil {
			writeError(w, http.StatusBadRequest, "invalid to, expected RFC3339")
			return
		}
	}

	var progs []model.Programme
	for _, p := range s.epgStore.Programmes() {
		if p.Channel != id {
			continue
		}
		if p.Stop.Before(from) || p.Start.After(to) {
			continue
		}
		progs = append(progs, toModelProgramme(p))
	}
	sort.Slice(progs, func(i, j int) bool { return progs[i].Start < progs[j].Start })

	writeJSON(w, http.StatusOK, struct {
		Channel    model.Channel     `json:"channel"`
		From       string            `json:"from"`
		To         string            `json:"to"`
		Programmes []model.Programme `json:"programmes"`
	}{ch, from.Format(time.RFC3339), to.Format(time.RFC3339), progs})
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func toModelProgramme(p epg.Programme) model.Programme {
	return model.Programme{
		Title:       p.Title,
		SubTitle:    p.SubTitle,
		Description: p.Desc,
		Icon:        p.Icon,
		Start:       p.Start.Format(time.RFC3339),
		Stop:        p.Stop.Format(time.RFC3339),
		Categories:  p.Categories,
		Rating:      p.Rating,
	}
}

func toModelEveningProgramme(p epg.Programme) model.EveningProgramme {
	return model.EveningProgramme{
		Title:       p.Title,
		SubTitle:    p.SubTitle,
		Description: p.Desc,
		Icon:        p.Icon,
		Start:       p.Start.Format(time.RFC3339),
		Stop:        p.Stop.Format(time.RFC3339),
		Categories:  p.Categories,
		Rating:      p.Rating,
	}
}
