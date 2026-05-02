package handler

import (
	"encoding/json"
	"net/http"

	chess "github.com/ajedrez4/ajedrez4-common"
	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

type drawHandler struct {
	sessions *store.SessionStore
	games    *store.GameStore
}

func newDrawHandler(sessions *store.SessionStore, games *store.GameStore) *drawHandler {
	return &drawHandler{sessions: sessions, games: games}
}

// draw handles POST /games/{id}/draw.
func (h *drawHandler) draw(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess := sessionFromContext(r.Context())

	g, ok := h.games.GetByID(id)
	if !ok {
		api.WriteError(w, http.StatusNotFound, domain.ErrGameNotFound, "Game not found")
		return
	}
	if g.Status != domain.GameStatusActive {
		api.WriteError(w, http.StatusConflict, domain.ErrGameNotActive, "Game is not active")
		return
	}
	if g.WhiteSessionID != sess.ID && g.BlackSessionID != sess.ID {
		api.WriteError(w, http.StatusForbidden, domain.ErrNotAParticipant, "You are not a participant in this game")
		return
	}

	callerColor := chess.White
	if g.BlackSessionID == sess.ID {
		callerColor = chess.Black
	}
	if g.ChessGame.ActivePlayer() != callerColor {
		api.WriteError(w, http.StatusForbidden, domain.ErrWrongTurn, "It is not your turn")
		return
	}

	var req api.ClaimDrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, domain.ErrInvalidDrawClaimType, "Invalid request body")
		return
	}

	var claimType chess.DrawClaimType
	switch req.ClaimType {
	case string(chess.DrawClaimThreefold):
		claimType = chess.DrawClaimThreefold
	case string(chess.DrawClaimFiftyMove):
		claimType = chess.DrawClaimFiftyMove
	default:
		api.WriteError(w, http.StatusBadRequest, domain.ErrInvalidDrawClaimType,
			"claim_type must be 'threefold_repetition' or 'fifty_move_rule'")
		return
	}

	result := g.ChessGame.ClaimDraw(claimType)

	// Rejected draw: do not update game state
	if !result.IsAccepted {
		reason := "Draw claim is not eligible at this point"
		if result.DrawDecision != nil {
			reason = result.DrawDecision.Reason
		}
		api.WriteJSON(w, http.StatusOK, api.DrawResponse{
			Accepted:        false,
			GameID:          id,
			Status:          string(domain.GameStatusActive),
			RejectionReason: &reason,
		})
		return
	}

	// Accepted draw: persist new state and mark finished
	newState := result.NewGameState
	_ = h.games.UpdateChessGame(id, newState, true)

	api.WriteJSON(w, http.StatusOK, api.DrawResponse{
		Accepted: true,
		GameID:   id,
		Status:   string(domain.GameStatusFinished),
	})
}
