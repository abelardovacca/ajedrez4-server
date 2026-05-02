package store

import (
	"errors"
	"sync"
	"time"

	chess "github.com/ajedrez4/ajedrez4-common"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
)

// Sentinel errors returned by GameStore operations.
var (
	ErrGameNotFound      = errors.New("game_not_found")
	ErrGameNotOpen       = errors.New("game_not_open")
	ErrCannotJoinOwnGame = errors.New("cannot_join_own_game")
	ErrCannotCancelGame  = errors.New("cannot_cancel_game")
)

// GameStore holds all games keyed by game ID.
type GameStore struct {
	mu    sync.RWMutex
	games map[string]*domain.Game
}

// NewGameStore returns an initialized GameStore.
func NewGameStore() *GameStore {
	return &GameStore{games: make(map[string]*domain.Game)}
}

// Create stores a new game.
func (s *GameStore) Create(whiteSessionID string) *domain.Game {
	id := newUUID()
	g := &domain.Game{
		ID:             id,
		Status:         domain.GameStatusOpen,
		WhiteSessionID: whiteSessionID,
		CreatedAt:      time.Now().UTC(),
	}
	s.mu.Lock()
	s.games[id] = g
	s.mu.Unlock()
	return g
}

// GetByID returns the game with the given ID.
func (s *GameStore) GetByID(id string) (*domain.Game, bool) {
	s.mu.RLock()
	g, ok := s.games[id]
	s.mu.RUnlock()
	return g, ok
}

// ListAll returns a snapshot of all games.
func (s *GameStore) ListAll() []*domain.Game {
	s.mu.RLock()
	out := make([]*domain.Game, 0, len(s.games))
	for _, g := range s.games {
		out = append(out, g)
	}
	s.mu.RUnlock()
	return out
}

// Delete removes a game by ID. Returns ErrGameNotFound if not present.
func (s *GameStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[id]; !ok {
		return ErrGameNotFound
	}
	delete(s.games, id)
	return nil
}

// JoinGame atomically transitions an open game to active.
// First-request-wins: only one concurrent caller will succeed.
func (s *GameStore) JoinGame(gameID, joinerSessionID string) (*domain.Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.games[gameID]
	if !ok {
		return nil, ErrGameNotFound
	}
	if g.WhiteSessionID == joinerSessionID {
		return nil, ErrCannotJoinOwnGame
	}
	if g.Status != domain.GameStatusOpen {
		return nil, ErrGameNotOpen
	}
	g.BlackSessionID = joinerSessionID
	g.Status = domain.GameStatusActive
	g.ChessGame = chess.NewGame(gameID)
	return g, nil
}

// UpdateChessGame persists a new chess state under write lock.
// If finished is true, the game status is set to finished.
// Should only be called after a move/resign/draw that changes state.
func (s *GameStore) UpdateChessGame(gameID string, newGame *chess.Game, finished bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.games[gameID]
	if !ok {
		return ErrGameNotFound
	}
	g.ChessGame = newGame
	if finished {
		g.Status = domain.GameStatusFinished
	}
	return nil
}
