package unit

import (
	"testing"

	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

func TestGameStore_CreateAndGet(t *testing.T) {
	gs := store.NewGameStore()
	g := gs.Create("session-white")
	if g.ID == "" {
		t.Error("expected non-empty game ID")
	}
	if g.Status != domain.GameStatusOpen {
		t.Errorf("expected open, got %q", g.Status)
	}
	got, ok := gs.GetByID(g.ID)
	if !ok {
		t.Fatal("expected to find game")
	}
	if got.ID != g.ID {
		t.Errorf("ID mismatch: %q vs %q", got.ID, g.ID)
	}
}

func TestGameStore_ListAll(t *testing.T) {
	gs := store.NewGameStore()
	gs.Create("s1")
	gs.Create("s2")
	all := gs.ListAll()
	if len(all) != 2 {
		t.Errorf("expected 2 games, got %d", len(all))
	}
}

func TestGameStore_Delete(t *testing.T) {
	gs := store.NewGameStore()
	g := gs.Create("s1")
	if err := gs.Delete(g.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, ok := gs.GetByID(g.ID)
	if ok {
		t.Error("expected game to be deleted")
	}
}

func TestGameStore_Delete_NotFound(t *testing.T) {
	gs := store.NewGameStore()
	if err := gs.Delete("nonexistent"); err == nil {
		t.Error("expected error deleting nonexistent game")
	}
}

func TestGameStore_JoinGame_Success(t *testing.T) {
	gs := store.NewGameStore()
	g := gs.Create("white-session")
	joined, err := gs.JoinGame(g.ID, "black-session")
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}
	if joined.Status != domain.GameStatusActive {
		t.Errorf("expected active, got %q", joined.Status)
	}
	if joined.BlackSessionID != "black-session" {
		t.Errorf("expected black-session, got %q", joined.BlackSessionID)
	}
	if joined.ChessGame == nil {
		t.Error("expected ChessGame to be initialized")
	}
}

func TestGameStore_JoinGame_DoubleJoin(t *testing.T) {
	gs := store.NewGameStore()
	g := gs.Create("white-session")
	_, _ = gs.JoinGame(g.ID, "black-session")
	_, err := gs.JoinGame(g.ID, "other-session")
	if err == nil {
		t.Error("expected error on double join")
	}
}

func TestGameStore_JoinGame_OwnGame(t *testing.T) {
	gs := store.NewGameStore()
	g := gs.Create("white-session")
	_, err := gs.JoinGame(g.ID, "white-session")
	if err == nil {
		t.Error("expected error joining own game")
	}
}

func TestGameStore_JoinGame_NotFound(t *testing.T) {
	gs := store.NewGameStore()
	_, err := gs.JoinGame("nonexistent", "some-session")
	if err == nil {
		t.Error("expected error for nonexistent game")
	}
}
