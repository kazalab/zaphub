package youtube

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStorePersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "channels.json")
	s := NewStore(path)

	if _, err := s.Add(Channel{ID: "UC1"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := s.Add(Channel{ID: "UC2", Name: "Chaîne 2"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if added, _ := s.Add(Channel{ID: "UC1"}); added {
		t.Fatal("expected duplicate to not be added")
	}

	s2 := NewStore(path)
	if err := s2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	channels := s2.Channels()
	if len(channels) != 2 || channels[0].ID != "UC1" || channels[1].Name != "Chaîne 2" {
		t.Fatalf("unexpected loaded channels: %+v", channels)
	}

	if removed, _ := s2.Remove("UC1"); !removed {
		t.Fatal("expected UC1 to be removed")
	}
	if removed, _ := s2.Remove("UC1"); removed {
		t.Fatal("expected UC1 to already be gone")
	}
	if len(s2.Channels()) != 1 {
		t.Fatalf("expected 1 channel, got %+v", s2.Channels())
	}
}

func TestStoreAllVideosSorted(t *testing.T) {
	s := NewStore("")
	t1, _ := time.Parse(time.RFC3339, "2026-08-11T00:00:00Z")
	t2, _ := time.Parse(time.RFC3339, "2026-08-12T00:00:00Z")
	s.UpdateVideos("UC1", []Video{
		{ID: "a", ChannelID: "UC1", ChannelName: "A", Published: t1},
		{ID: "b", ChannelID: "UC1", ChannelName: "A", Published: t2},
	})
	s.UpdateVideos("UC2", []Video{
		{ID: "c", ChannelID: "UC2", ChannelName: "B", Published: t2.Add(-time.Hour)},
	})

	all := s.AllVideos()
	if len(all) != 3 {
		t.Fatalf("expected 3 videos, got %d", len(all))
	}
	if all[0].ID != "b" || all[1].ID != "c" || all[2].ID != "a" {
		t.Fatalf("unexpected sort order: %+v", all)
	}
}
