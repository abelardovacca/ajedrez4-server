package unit

import (
	"testing"

	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

func TestSessionStore_CreateReturnsUniqueIDs(t *testing.T) {
	s := store.NewSessionStore()
	a := s.Create("Alice")
	b := s.Create("Alice")
	if a.ID == b.ID {
		t.Errorf("expected unique IDs, got both %q", a.ID)
	}
}

func TestSessionStore_GetByID_Found(t *testing.T) {
	s := store.NewSessionStore()
	created := s.Create("Bob")
	got, ok := s.GetByID(created.ID)
	if !ok {
		t.Fatal("expected to find session, got ok=false")
	}
	if got.DisplayName != "Bob" {
		t.Errorf("expected DisplayName=Bob, got %q", got.DisplayName)
	}
}

func TestSessionStore_GetByID_NotFound(t *testing.T) {
	s := store.NewSessionStore()
	_, ok := s.GetByID("nonexistent")
	if ok {
		t.Error("expected ok=false for missing session")
	}
}

func TestSessionStore_DisplayNamePreserved(t *testing.T) {
	s := store.NewSessionStore()
	sess := s.Create("Chess Master")
	got, _ := s.GetByID(sess.ID)
	if got.DisplayName != "Chess Master" {
		t.Errorf("got %q", got.DisplayName)
	}
}
