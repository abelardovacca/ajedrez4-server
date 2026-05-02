package handler

import (
	chess "github.com/ajedrez4/ajedrez4-common"
	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

// lookupDisplayName resolves a session ID to a display name via the store.
// Returns empty string if not found (should not happen for valid game participants).
func lookupDisplayName(sessions *store.SessionStore, sessionID string) string {
	sess, ok := sessions.GetByID(sessionID)
	if !ok {
		return ""
	}
	return sess.DisplayName
}

// toSummary converts a domain.Game to a GameSummaryResponse.
func toSummary(g *domain.Game, sessions *store.SessionStore) api.GameSummaryResponse {
	var blackPlayer *string
	if g.BlackSessionID != "" {
		name := lookupDisplayName(sessions, g.BlackSessionID)
		blackPlayer = &name
	}
	return api.GameSummaryResponse{
		GameID:      g.ID,
		Status:      string(g.Status),
		WhitePlayer: lookupDisplayName(sessions, g.WhiteSessionID),
		BlackPlayer: blackPlayer,
		CreatedAt:   g.CreatedAt,
	}
}

// toDetail converts a domain.Game to a GameDetailResponse.
func toDetail(g *domain.Game, sessions *store.SessionStore) api.GameDetailResponse {
	var blackPlayer *string
	if g.BlackSessionID != "" {
		name := lookupDisplayName(sessions, g.BlackSessionID)
		blackPlayer = &name
	}

	var fenPtr *string
	var activeColorPtr *string
	var moveHistory []string
	var outcome *api.OutcomeResponse

	if g.ChessGame != nil {
		fen := g.ChessGame.ExportFEN()
		fenPtr = &fen
		moveHistory = g.ChessGame.MoveHistory()
		if g.Status == domain.GameStatusActive {
			ac := g.ChessGame.ActivePlayer().String()
			activeColorPtr = &ac
		}
		if o := g.ChessGame.Outcome(); o != nil {
			outcome = toOutcome(o)
		}
	}
	if moveHistory == nil {
		moveHistory = []string{}
	}

	return api.GameDetailResponse{
		GameID:      g.ID,
		Status:      string(g.Status),
		WhitePlayer: lookupDisplayName(sessions, g.WhiteSessionID),
		BlackPlayer: blackPlayer,
		CreatedAt:   g.CreatedAt,
		ActiveColor: activeColorPtr,
		FEN:         fenPtr,
		MoveHistory: moveHistory,
		Outcome:     outcome,
	}
}

// toOutcome converts a chess.GameOutcome to an OutcomeResponse.
func toOutcome(o *chess.GameOutcome) *api.OutcomeResponse {
	result := outcomeResult(o)
	var winner *string
	var drawReason *string
	if o.Winner != chess.NoColor {
		w := o.Winner.String()
		winner = &w
	}
	if o.DrawReason != "" {
		drawReason = &o.DrawReason
	}
	return &api.OutcomeResponse{
		Result:     result,
		Winner:     winner,
		DrawReason: drawReason,
	}
}

func outcomeResult(o *chess.GameOutcome) string {
	switch o.Result {
	case chess.ResultWhiteWins:
		return "white_wins"
	case chess.ResultBlackWins:
		return "black_wins"
	case chess.ResultDraw:
		return "draw"
	default:
		return "in_progress"
	}
}
