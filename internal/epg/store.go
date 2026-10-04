package epg

import (
	"sync"
	"time"

	"github.com/kazalab/zaphub/internal/model"
)

type Store struct {
	mu          sync.RWMutex
	channels    []model.Channel
	byID        map[string]model.Channel
	programmes  []Programme
	lastRefresh time.Time
	lastSuccess time.Time
	lastError   string
}

func NewStore() *Store {
	return &Store{byID: make(map[string]model.Channel)}
}

func (s *Store) Update(channels []model.Channel, programmes []Programme) {
	byID := make(map[string]model.Channel, len(channels))
	for _, c := range channels {
		byID[c.ID] = c
	}
	s.mu.Lock()
	s.channels = channels
	s.byID = byID
	s.programmes = programmes
	s.lastRefresh = time.Now()
	s.lastSuccess = time.Now()
	s.lastError = ""
	s.mu.Unlock()
}

func (s *Store) SetError(err error) {
	s.mu.Lock()
	s.lastRefresh = time.Now()
	if err != nil {
		s.lastError = err.Error()
	}
	s.mu.Unlock()
}

func (s *Store) Channels() []model.Channel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Channel, len(s.channels))
	copy(out, s.channels)
	return out
}

func (s *Store) Channel(id string) (model.Channel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.byID[id]
	return c, ok
}

func (s *Store) Programmes() []Programme {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Programme, len(s.programmes))
	copy(out, s.programmes)
	return out
}

// Now returns, per channel, the programme currently airing at the given time.
func (s *Store) Now(now time.Time) map[string]Programme {
	s.mu.RLock()
	progs := s.programmes
	s.mu.RUnlock()

	current := make(map[string]Programme)
	for _, p := range progs {
		if now.Before(p.Start) || !now.Before(p.Stop) {
			continue
		}
		if cur, ok := current[p.Channel]; !ok || p.Start.After(cur.Start) {
			current[p.Channel] = p
		}
	}
	return current
}

func (s *Store) Status() (lastRefresh, lastSuccess time.Time, lastError string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRefresh, s.lastSuccess, s.lastError
}
