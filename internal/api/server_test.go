package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/kazalab/zaphub/internal/cinema"
	"github.com/kazalab/zaphub/internal/epg"
	"github.com/kazalab/zaphub/internal/model"
	"github.com/kazalab/zaphub/internal/youtube"
)

func newTestServer(t *testing.T) (*Server, *epg.Store, *cinema.Store, *youtube.Store) {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "epg", "testdata", "epg_sample.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	channels, programmes, err := epg.Parse(f)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	epgStore := epg.NewStore()
	epgStore.Update(channels, programmes)

	cinemaStore := cinema.NewStore("test")
	cinemaStore.Update([]model.CinemaRelease{
		{ID: "1", Title: "Film Test", ReleaseDate: "2026-09-02"},
	})

	ytStore := youtube.NewStore("")
	pub1, _ := time.Parse(time.RFC3339, "2026-08-11T18:08:30Z")
	pub2, _ := time.Parse(time.RFC3339, "2026-08-10T09:00:00Z")
	if _, err := ytStore.Add(youtube.Channel{ID: "UC1", Name: "Chaîne 1"}); err != nil {
		t.Fatalf("add channel UC1: %v", err)
	}
	if _, err := ytStore.Add(youtube.Channel{ID: "UC2", Name: "Chaîne 2"}); err != nil {
		t.Fatalf("add channel UC2: %v", err)
	}
	ytStore.UpdateVideos("UC1", []youtube.Video{
		{ID: "a", Title: "Vidéo A", Link: "https://youtube.com/watch?v=a", Published: pub1, ChannelID: "UC1", ChannelName: "Chaîne 1"},
	})
	ytStore.UpdateVideos("UC2", []youtube.Video{
		{ID: "b", Title: "Vidéo B", Link: "https://youtube.com/watch?v=b", Published: pub2, ChannelID: "UC2", ChannelName: "Chaîne 2"},
	})

	return New(Options{EPGStore: epgStore, CinemaStore: cinemaStore, YTStore: ytStore}), epgStore, cinemaStore, ytStore
}

func doRequest(t *testing.T, h http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rdr)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestEPGChannels(t *testing.T) {
	s, _, _, _ := newTestServer(t)
	rec := doRequest(t, s.Handler(), http.MethodGet, "/api/epg/channels", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var channels []model.Channel
	if err := json.Unmarshal(rec.Body.Bytes(), &channels); err != nil {
		t.Fatal(err)
	}
	if len(channels) != 3 || channels[0].ID != "TF1.fr" {
		t.Fatalf("unexpected channels: %+v", channels)
	}
}

func TestEPGNow(t *testing.T) {
	s, _, _, _ := newTestServer(t)

	rec := doRequest(t, s.Handler(), http.MethodGet, "/api/epg/now", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var entries []nowEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/epg/now?channel=TF1.fr", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var single nowEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &single); err != nil {
		t.Fatal(err)
	}
	if single.ChannelName != "TF1" {
		t.Fatalf("unexpected entry: %+v", single)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/epg/now?channel=NOPE.fr", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestEPGEvening(t *testing.T) {
	s, _, _, _ := newTestServer(t)
	rec := doRequest(t, s.Handler(), http.MethodGet, "/api/epg/evening?date=2026-08-11", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var resp model.EveningResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Date != "2026-08-11" {
		t.Fatalf("unexpected date: %s", resp.Date)
	}
	if len(resp.Channels) != 3 {
		t.Fatalf("expected 3 channels, got %d", len(resp.Channels))
	}
	tf1 := resp.Channels[0]
	if tf1.ChannelID != "TF1.fr" || len(tf1.Programmes) != 2 {
		t.Fatalf("unexpected TF1 evening: %+v", tf1)
	}
	if tf1.Programmes[0].Title != "Film du soir TF1" || tf1.Programmes[1].Title != "Deuxième programme TF1" {
		t.Fatalf("unexpected TF1 programmes: %+v", tf1)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/epg/evening?date=2026-08-11&channel=France2.fr", nil)
	var single model.EveningResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &single); err != nil {
		t.Fatal(err)
	}
	if len(single.Channels) != 1 || single.Channels[0].ChannelID != "France2.fr" {
		t.Fatalf("unexpected filtered result: %+v", single)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/epg/evening?date=2026-08-11&channel=NOPE.fr", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/epg/evening?date=11/08/2026", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestEPGEveningMinDuration(t *testing.T) {
	s, _, _, _ := newTestServer(t)
	rec := doRequest(t, s.Handler(), http.MethodGet, "/api/epg/evening?date=2026-08-11", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var resp model.EveningResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	paris := time.FixedZone("CEST", 2*3600)
	for _, ch := range resp.Channels {
		for _, p := range ch.Programmes {
			start, err := time.Parse(time.RFC3339, p.Start)
			if err != nil {
				t.Fatal(err)
			}
			stop, err := time.Parse(time.RFC3339, p.Stop)
			if err != nil {
				t.Fatal(err)
			}
			if d := stop.Sub(start); d < 40*time.Minute {
				t.Fatalf("%s %q lasts %s, expected >= 40m", ch.ChannelID, p.Title, d)
			}
			sIn := start.In(paris)
			minutes := sIn.Hour()*60 + sIn.Minute()
			if minutes < 20*60+50 || minutes > 23*60+50 {
				t.Fatalf("%s %q starts at %s, outside [20:50, 23:50]", ch.ChannelID, p.Title, sIn.Format("15:04"))
			}
		}
	}
}

func TestEPGProgrammes(t *testing.T) {
	s, _, _, _ := newTestServer(t)
	rec := doRequest(t, s.Handler(), http.MethodGet,
		"/api/epg/programmes?channel=TF1.fr&from=2026-08-11T00:00:00%2B02:00&to=2026-08-12T00:00:00%2B02:00", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Programmes []model.Programme `json:"programmes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Programmes) != 3 {
		t.Fatalf("expected 3 programmes, got %d", len(resp.Programmes))
	}
	if resp.Programmes[0].Title != "Avant Soirée TF1" {
		t.Fatalf("unexpected first programme: %+v", resp.Programmes[0])
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/epg/programmes", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCinemaReleases(t *testing.T) {
	s, _, _, _ := newTestServer(t)

	rec := doRequest(t, s.Handler(), http.MethodGet, "/api/cinema/releases", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var releases []model.CinemaRelease
	if err := json.Unmarshal(rec.Body.Bytes(), &releases); err != nil {
		t.Fatal(err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
	if releases[0].Title != "Film Test" {
		t.Fatalf("unexpected release: %+v", releases[0])
	}
}

func TestYTVideos(t *testing.T) {
	s, _, _, _ := newTestServer(t)

	rec := doRequest(t, s.Handler(), http.MethodGet, "/api/youtube/videos", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var videos []model.Video
	if err := json.Unmarshal(rec.Body.Bytes(), &videos); err != nil {
		t.Fatal(err)
	}
	if len(videos) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(videos))
	}
	if videos[0].ID != "a" {
		t.Fatalf("expected video a first (newest), got %+v", videos[0])
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/youtube/videos?channel=UC1", nil)
	if err := json.Unmarshal(rec.Body.Bytes(), &videos); err != nil {
		t.Fatal(err)
	}
	if len(videos) != 1 || videos[0].ID != "a" {
		t.Fatalf("unexpected filtered videos: %+v", videos)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/youtube/videos?channel=NOPE", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestYoutubeCRUD(t *testing.T) {
	s, _, _, _ := newTestServer(t)

	rec := doRequest(t, s.Handler(), http.MethodPost, "/api/youtube/channels", []byte(`{"channel_id":"UCnew"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, s.Handler(), http.MethodPost, "/api/youtube/channels", []byte(`{"channel_id":"UCnew"}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on duplicate, got %d", rec.Code)
	}

	rec = doRequest(t, s.Handler(), http.MethodPost, "/api/youtube/channels", []byte(`{}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	rec = doRequest(t, s.Handler(), http.MethodDelete, "/api/youtube/channels/UCnew", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec = doRequest(t, s.Handler(), http.MethodDelete, "/api/youtube/channels/UCnew", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHealth(t *testing.T) {
	s, _, _, _ := newTestServer(t)
	rec := doRequest(t, s.Handler(), http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}

	cinema, ok := body["cinema"].(map[string]any)
	if !ok {
		t.Fatal("missing cinema block in health")
	}
	if cinema["provider"] != "test" {
		t.Fatalf("unexpected cinema provider: %v", cinema["provider"])
	}
}

func TestWebUI(t *testing.T) {
	_, epgStore, cinemaStore, ytStore := newTestServer(t)

	web := fstest.MapFS{
		"index.html":  &fstest.MapFile{Data: []byte("<!DOCTYPE html><title>ZapHub</title>")},
		"status.html": &fstest.MapFile{Data: []byte("<!DOCTYPE html><title>Statut</title>")},
		"style.css":   &fstest.MapFile{Data: []byte("body{}")},
		"app.js":      &fstest.MapFile{Data: []byte("console.log('hi')")},
	}
	s := New(Options{EPGStore: epgStore, CinemaStore: cinemaStore, YTStore: ytStore, WebFS: web})

	rec := doRequest(t, s.Handler(), http.MethodGet, "/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("unexpected content-type: %s", ct)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/status.html", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d for /status.html", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("unexpected content-type for /status.html: %s", ct)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/style.css", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d for /style.css", rec.Code)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/nope", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /nope, got %d", rec.Code)
	}

	rec = doRequest(t, s.Handler(), http.MethodGet, "/api/epg/channels", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("api route shadowed by static: status %d", rec.Code)
	}
}
