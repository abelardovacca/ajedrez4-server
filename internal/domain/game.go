package domain

import (
	"time"

	chess "github.com/ajedrez4/ajedrez4-common"
)

// GameStatus represents the lifecycle state of a game.
type GameStatus string

const (
	GameStatusOpen     GameStatus = "open"
	GameStatusActive   GameStatus = "active"
	GameStatusFinished GameStatus = "finished"
)

// Game represents a chess match lobby and its associated game state.
type Game struct {
	ID             string
	Status         GameStatus
	WhiteSessionID string
	BlackSessionID string // empty when Status == open
	ChessGame      *chess.Game // nil when Status == open
	CreatedAt      time.Time
}
