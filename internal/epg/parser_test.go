package epg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "epg_sample.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	channels, programmes, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(channels) != 3 {
		t.Fatalf("expected 3 channels, got %d", len(channels))
	}
	if len(programmes) != 6 {
		t.Fatalf("expected 6 programmes, got %d", len(programmes))
	}

	tf1 := channels[0]
	if tf1.ID != "TF1.fr" || tf1.Name != "TF1" || tf1.Icon != "https://example.com/tf1.png" {
		t.Fatalf("unexpected channel: %+v", tf1)
	}

	p := programmes[1]
	if p.Channel != "TF1.fr" {
		t.Fatalf("expected TF1.fr, got %s", p.Channel)
	}
	if p.Title != "Film du soir TF1" || p.SubTitle != "Épisode 1" {
		t.Fatalf("unexpected title: %+v", p)
	}
	if p.Desc != "Un synopsis de test." {
		t.Fatalf("unexpected desc: %q", p.Desc)
	}
	if p.Icon != "https://example.com/film-du-soir.jpg" {
		t.Fatalf("unexpected icon: %q", p.Icon)
	}
	if len(p.Categories) != 2 || p.Categories[0] != "Cinéma" {
		t.Fatalf("unexpected categories: %v", p.Categories)
	}
	if p.Rating != "Tout public" {
		t.Fatalf("unexpected rating: %q", p.Rating)
	}
	wantStart := mustTime(t, "20260811210000 +0200")
	if !p.Start.Equal(wantStart) {
		t.Fatalf("expected start %v, got %v", wantStart, p.Start)
	}
}

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(xmlTimeLayout, s)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return ts
}
