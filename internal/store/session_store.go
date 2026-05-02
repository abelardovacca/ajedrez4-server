package store

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/ajedrez4/ajedrez4-server/internal/domain"
)

// SessionStore holds all active sessions keyed by session ID.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

// NewSessionStore returns an initialized SessionStore.
func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]*domain.Session)}
}

// Create generates a new UUID v4 for the session and stores it.
func (s *SessionStore) Create(displayName string) *domain.Session {
	id := newUUID()
	session := &domain.Session{
		ID:          id,
		DisplayName: displayName,
		CreatedAt:   time.Now().UTC(),
	}
	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()
	return session
}

// GetByID returns the session with the given ID, or (nil, false) if not found.
func (s *SessionStore) GetByID(id string) (*domain.Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	return sess, ok
}

// newUUID generates a random UUID v4.
func newUUID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(fmt.Sprintf("crypto/rand unavailable: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
