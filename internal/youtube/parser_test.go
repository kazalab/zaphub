package youtube

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "yt_sample.xml"))
	if err != nil {
		t.Fatal(err)
	}
	videos, err := ParseBytes(data, "UCabc", "")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(videos) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(videos))
	}

	v := videos[0]
	if v.ID != "VID1" || v.Title != "Première vidéo" {
		t.Fatalf("unexpected video: %+v", v)
	}
	if v.Link != "https://www.youtube.com/watch?v=VID1" {
		t.Fatalf("unexpected link: %q", v.Link)
	}
	if v.Thumbnail != "https://i.ytimg.com/vi/VID1/hqdefault.jpg" {
		t.Fatalf("unexpected thumbnail: %q", v.Thumbnail)
	}
	if v.Likes != 5100 {
		t.Fatalf("unexpected likes: %d", v.Likes)
	}
	if v.Duration != "320" {
		t.Fatalf("unexpected duration: %q", v.Duration)
	}
	if v.ChannelID != "UCabc" {
		t.Fatalf("unexpected channel id: %q", v.ChannelID)
	}
	if v.ChannelName != "Chaîne Test" {
		t.Fatalf("unexpected channel name resolved from feed: %q", v.ChannelName)
	}
	wantPub, _ := time.Parse(time.RFC3339, "2026-08-11T18:08:30+00:00")
	if !v.Published.Equal(wantPub) {
		t.Fatalf("unexpected published: %v", v.Published)
	}
}
