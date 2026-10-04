package epg

import (
	"os"
	"path/filepath"
	"testing"
)

func loadFixture(t *testing.T) ([]Programme, error) {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "epg_sample.xml"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	_, progs, err := Parse(f)
	return progs, err
}

func TestEveningProgrammes(t *testing.T) {
	progs, err := loadFixture(t)
	if err != nil {
		t.Fatal(err)
	}
	date := mustTime(t, "20260811000000 +0200")

	got := EveningProgrammes(progs, date)

	tf1 := got["TF1.fr"]
	if len(tf1) != 2 {
		t.Fatalf("TF1.fr: expected 2 evening programmes, got %d", len(tf1))
	}
	if tf1[0].Title != "Film du soir TF1" || tf1[1].Title != "Deuxième programme TF1" {
		t.Fatalf("TF1.fr unexpected selection: %+v", tf1)
	}

	fr2 := got["France2.fr"]
	if len(fr2) != 2 || fr2[0].Title != "Soirée France 2" || fr2[1].Title != "Suite France 2" {
		t.Fatalf("France2.fr unexpected selection: %+v", fr2)
	}

	if m6 := got["M6.fr"]; len(m6) != 0 {
		t.Fatalf("M6.fr: expected no evening programmes, got %+v", m6)
	}
}

func TestEveningProgrammesOtherDay(t *testing.T) {
	progs, err := loadFixture(t)
	if err != nil {
		t.Fatal(err)
	}
	date := mustTime(t, "20260812000000 +0200")
	got := EveningProgrammes(progs, date)
	for _, ch := range []string{"TF1.fr", "France2.fr", "M6.fr"} {
		if len(got[ch]) != 0 {
			t.Fatalf("%s: expected no evening programmes on 2026-08-12, got %+v", ch, got[ch])
		}
	}
}

func TestEveningProgrammesCriteria(t *testing.T) {
	mk := func(ch, start, stop, title string) Programme {
		return Programme{
			Channel: ch,
			Start:   mustTime(t, start),
			Stop:    mustTime(t, stop),
			Title:   title,
		}
	}
	progs := []Programme{
		// start window boundaries
		mk("CH1.fr", "20260811204900 +0200", "20260811214900 +0200", "début 20:49 exclu"),
		mk("CH1.fr", "20260811205000 +0200", "20260811215000 +0200", "début pile 20:50"),
		mk("CH1.fr", "20260811235000 +0200", "20260812005000 +0200", "début pile 23:50"),
		mk("CH1.fr", "20260811235100 +0200", "20260812015100 +0200", "début 23:51 exclu"),
		// duration boundary (all start 21:00, 60 min span)
		mk("CH2.fr", "20260811210000 +0200", "20260811213900 +0200", "39 min exclu"),
		mk("CH2.fr", "20260811213900 +0200", "20260811222000 +0200", "pile 41 min"),
		mk("CH2.fr", "20260811222000 +0200", "20260811230000 +0200", "40 min pile"),
		// no limit: three long programmes in the window all returned
		mk("CH3.fr", "20260811210000 +0200", "20260811230000 +0200", "Long 1"),
		mk("CH3.fr", "20260811230000 +0200", "20260811235000 +0200", "Long 2"),
		mk("CH3.fr", "20260811235000 +0200", "20260812005000 +0200", "Long 3"),
	}
	date := mustTime(t, "20260811000000 +0200")

	got := EveningProgrammes(progs, date)

	ch1 := got["CH1.fr"]
	if len(ch1) != 2 || ch1[0].Title != "début pile 20:50" || ch1[1].Title != "début pile 23:50" {
		t.Fatalf("CH1.fr: start window not respected, got %+v", ch1)
	}

	ch2 := got["CH2.fr"]
	if len(ch2) != 2 || ch2[0].Title != "pile 41 min" || ch2[1].Title != "40 min pile" {
		t.Fatalf("CH2.fr: duration boundary not respected, got %+v", ch2)
	}

	ch3 := got["CH3.fr"]
	if len(ch3) != 3 || ch3[0].Title != "Long 1" || ch3[1].Title != "Long 2" || ch3[2].Title != "Long 3" {
		t.Fatalf("CH3.fr: no-limit selection not respected, got %+v", ch3)
	}
}

func TestStoreNow(t *testing.T) {
	progs, err := loadFixture(t)
	if err != nil {
		t.Fatal(err)
	}
	s := NewStore()
	s.Update(nil, progs)

	current := s.Now(mustTime(t, "20260811213000 +0200"))
	if p, ok := current["TF1.fr"]; !ok || p.Title != "Film du soir TF1" {
		t.Fatalf("expected Film du soir TF1 at 21:30, got %+v", p)
	}
	if p, ok := current["M6.fr"]; ok {
		t.Fatalf("M6.fr should have no current programme at 21:30, got %+v", p)
	}

	// Crossing midnight: 22:40-00:05 should be active at 23:30
	current = s.Now(mustTime(t, "20260811233000 +0200"))
	if p, ok := current["TF1.fr"]; !ok || p.Title != "Deuxième programme TF1" {
		t.Fatalf("expected Deuxième programme TF1 at 23:30, got %+v", p)
	}
}
