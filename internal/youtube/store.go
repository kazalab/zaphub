package youtube

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Channel struct {
	ID   string `json:"channel_id"`
	Name string `json:"name,omitempty"`
}

type Store struct {
	mu          sync.RWMutex
	path        string
	channels    []Channel
	videos      map[string][]Video
	lastRefresh time.Time
	lastError   string
}

func NewStore(path string) *Store {
	return &Store{path: path, videos: make(map[string][]Video)}
}

func (s *Store) Load() error {
	if s.path == "" {
		return nil
	}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var p struct {
		Channels []Channel `json:"channels"`
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	s.mu.Lock()
	s.channels = p.Channels
	s.mu.Unlock()
	return nil
}

func (s *Store) Save() error {
	if s.path == "" {
		return nil
	}
	s.mu.RLock()
	p := struct {
		Channels []Channel `json:"channels"`
	}{Channels: s.channels}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Add(ch Channel) (bool, error) {
	s.mu.Lock()
	for _, c := range s.channels {
		if c.ID == ch.ID {
			s.mu.Unlock()
			return false, nil
		}
	}
	s.channels = append(s.channels, ch)
	s.mu.Unlock()
	return true, s.Save()
}

func (s *Store) SetChannel(ch Channel) error {
	s.mu.Lock()
	for i, c := range s.channels {
		if c.ID == ch.ID {
			s.channels[i] = ch
			s.mu.Unlock()
			return s.Save()
		}
	}
	s.mu.Unlock()
	return nil
}

func (s *Store) Remove(id string) (bool, error) {
	s.mu.Lock()
	for i, c := range s.channels {
		if c.ID == id {
			s.channels = append(s.channels[:i], s.channels[i+1:]...)
			delete(s.videos, id)
			s.mu.Unlock()
			return true, s.Save()
		}
	}
	s.mu.Unlock()
	return false, nil
}

func (s *Store) Get(id string) (Channel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.channels {
		if c.ID == id {
			return c, true
		}
	}
	return Channel{}, false
}

func (s *Store) Channels() []Channel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Channel, len(s.channels))
	copy(out, s.channels)
	return out
}

func (s *Store) UpdateVideos(id string, videos []Video) {
	s.mu.Lock()
	if len(videos) > 0 {
		s.videos[id] = videos
	} else {
		delete(s.videos, id)
	}
	s.lastRefresh = time.Now()
	s.lastError = ""
	s.mu.Unlock()
}

func (s *Store) Videos(id string) []Video {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Video, len(s.videos[id]))
	copy(out, s.videos[id])
	return out
}

func (s *Store) AllVideos() []Video {
	s.mu.RLock()
	var out []Video
	for _, vs := range s.videos {
		out = append(out, vs...)
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Published.After(out[j].Published) })
	return out
}

func (s *Store) SetError(err error) {
	s.mu.Lock()
	s.lastRefresh = time.Now()
	if err != nil {
		s.lastError = err.Error()
	}
	s.mu.Unlock()
}

func (s *Store) Status() (lastRefresh time.Time, lastError string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRefresh, s.lastError
}
