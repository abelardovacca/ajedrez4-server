package handler

import (
	"errors"
	"net/http"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

type gameHandler struct {
	sessions *store.SessionStore
	games    *store.GameStore
}

func newGameHandler(sessions *store.SessionStore, games *store.GameStore) *gameHandler {
	return &gameHandler{sessions: sessions, games: games}
}

// list handles GET /games.
func (h *gameHandler) list(w http.ResponseWriter, r *http.Request) {
	all := h.games.ListAll()
	summaries := make([]api.GameSummaryResponse, 0, len(all))
	for _, g := range all {
		summaries = append(summaries, toSummary(g, h.sessions))
	}
	api.WriteJSON(w, http.StatusOK, api.GamesListResponse{Games: summaries})
}

// create handles POST /games.
func (h *gameHandler) create(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromContext(r.Context())
	g := h.games.Create(sess.ID)
	api.WriteJSON(w, http.StatusCreated, toSummary(g, h.sessions))
}

// detail handles GET /games/{id}.
func (h *gameHandler) detail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, ok := h.games.GetByID(id)
	if !ok {
		api.WriteError(w, http.StatusNotFound, domain.ErrGameNotFound, "Game not found")
		return
	}
	api.WriteJSON(w, http.StatusOK, toDetail(g, h.sessions))
}

// cancel handles DELETE /games/{id}.
func (h *gameHandler) cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess := sessionFromContext(r.Context())

	g, ok := h.games.GetByID(id)
	if !ok {
		api.WriteError(w, http.StatusNotFound, domain.ErrGameNotFound, "Game not found")
		return
	}
	if g.WhiteSessionID != sess.ID || g.Status != domain.GameStatusOpen {
		api.WriteError(w, http.StatusForbidden, domain.ErrCannotCancelGame,
			"Only the creator may cancel an open game")
		return
	}
	if err := h.games.Delete(id); err != nil {
		if errors.Is(err, store.ErrGameNotFound) {
			api.WriteError(w, http.StatusNotFound, domain.ErrGameNotFound, "Game not found")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "internal_error", "Unexpected error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
