package handler

import (
	"net/http"

	chess "github.com/ajedrez4/ajedrez4-common"
	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

type resignHandler struct {
	sessions *store.SessionStore
	games    *store.GameStore
}

func newResignHandler(sessions *store.SessionStore, games *store.GameStore) *resignHandler {
	return &resignHandler{sessions: sessions, games: games}
}

// resign handles POST /games/{id}/resign.
func (h *resignHandler) resign(w http.ResponseWriter, r *http.Request) {
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

	resignColor := chess.White
	if g.BlackSessionID == sess.ID {
		resignColor = chess.Black
	}

	result := g.ChessGame.Resign(resignColor)
	newState := result.NewGameState
	_ = h.games.UpdateChessGame(id, newState, true)

	outcome := toOutcome(newState.Outcome())
	api.WriteJSON(w, http.StatusOK, api.ResignResponse{
		GameID:  id,
		Status:  string(domain.GameStatusFinished),
		Outcome: *outcome,
	})
}
