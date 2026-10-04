package cinema

import (
	"sync"
	"time"

	"github.com/kazalab/zaphub/internal/model"
)

type Store struct {
	mu          sync.RWMutex
	provider    string
	releases    []model.CinemaRelease
	lastRefresh time.Time
	lastSuccess time.Time
	lastError   string
}

func NewStore(provider string) *Store {
	return &Store{provider: provider}
}

func (s *Store) Update(releases []model.CinemaRelease) {
	s.mu.Lock()
	s.releases = releases
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

func (s *Store) Releases() []model.CinemaRelease {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.CinemaRelease, len(s.releases))
	copy(out, s.releases)
	return out
}

func (s *Store) Status() (lastRefresh, lastSuccess time.Time, lastError string, provider string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRefresh, s.lastSuccess, s.lastError, s.provider
}
