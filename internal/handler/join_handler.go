package handler

import (
	"errors"
	"net/http"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

type joinHandler struct {
	sessions *store.SessionStore
	games    *store.GameStore
}

func newJoinHandler(sessions *store.SessionStore, games *store.GameStore) *joinHandler {
	return &joinHandler{sessions: sessions, games: games}
}

// join handles POST /games/{id}/join.
func (h *joinHandler) join(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess := sessionFromContext(r.Context())

	g, err := h.games.JoinGame(id, sess.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrGameNotFound):
			api.WriteError(w, http.StatusNotFound, domain.ErrGameNotFound, "Game not found")
		case errors.Is(err, store.ErrCannotJoinOwnGame):
			api.WriteError(w, http.StatusForbidden, domain.ErrCannotJoinOwnGame,
				"You cannot join your own game")
		case errors.Is(err, store.ErrGameNotOpen):
			api.WriteError(w, http.StatusConflict, domain.ErrGameNotOpen,
				"Game is not open for joining")
		default:
			api.WriteError(w, http.StatusInternalServerError, "internal_error", "Unexpected error")
		}
		return
	}
	api.WriteJSON(w, http.StatusOK, toDetail(g, h.sessions))
}
