package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

type sessionHandler struct {
	sessions *store.SessionStore
}

func newSessionHandler(sessions *store.SessionStore) *sessionHandler {
	return &sessionHandler{sessions: sessions}
}

func (h *sessionHandler) create(w http.ResponseWriter, r *http.Request) {
	var req api.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, domain.ErrInvalidDisplayName, "Invalid request body")
		return
	}

	if !domain.ValidateDisplayName(req.DisplayName) {
		api.WriteError(w, http.StatusBadRequest, domain.ErrInvalidDisplayName,
			"Display name must be 1–32 printable characters")
		return
	}

	sess := h.sessions.Create(req.DisplayName)
	api.WriteJSON(w, http.StatusCreated, api.SessionResponse{
		SessionID:   sess.ID,
		DisplayName: sess.DisplayName,
		CreatedAt:   sess.CreatedAt,
	})
}
