package handler

import (
	"encoding/json"
	"net/http"

	chess "github.com/ajedrez4/ajedrez4-common"
	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

type moveHandler struct {
	sessions *store.SessionStore
	games    *store.GameStore
}

func newMoveHandler(sessions *store.SessionStore, games *store.GameStore) *moveHandler {
	return &moveHandler{sessions: sessions, games: games}
}

// move handles POST /games/{id}/moves.
func (h *moveHandler) move(w http.ResponseWriter, r *http.Request) {
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

	var req api.SubmitMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Notation == "" {
		api.WriteError(w, http.StatusBadRequest, domain.ErrInvalidNotation, "Invalid move notation")
		return
	}

	result := g.ChessGame.SubmitMove(req.Notation)
	switch result.Status {
	case chess.MoveInvalidNotation:
		api.WriteError(w, http.StatusBadRequest, domain.ErrInvalidNotation, result.RejectionReason.Message)
		return
	case chess.MoveIllegal:
		api.WriteError(w, http.StatusUnprocessableEntity, domain.ErrIllegalMove, result.RejectionReason.Message)
		return
	case chess.MoveGameAlreadyFinished:
		api.WriteError(w, http.StatusConflict, domain.ErrGameNotActive, "Game is already finished")
		return
	}

	newState := result.NewGameState
	finished := newState.Status() != chess.InProgress
	_ = h.games.UpdateChessGame(id, newState, finished)

	var activeColor *string
	if !finished {
		ac := newState.ActivePlayer().String()
		activeColor = &ac
	}
	var outcome *api.OutcomeResponse
	if o := newState.Outcome(); o != nil {
		outcome = toOutcome(o)
	}

	api.WriteJSON(w, http.StatusOK, api.MoveResponse{
		Accepted:    true,
		GameID:      id,
		Status:      statusString(finished),
		ActiveColor: activeColor,
		FEN:         newState.ExportFEN(),
		MoveHistory: newState.MoveHistory(),
		Outcome:     outcome,
	})
}

func statusString(finished bool) string {
	if finished {
		return string(domain.GameStatusFinished)
	}
	return string(domain.GameStatusActive)
}
