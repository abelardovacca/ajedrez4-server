package handler

import (
	"net/http"

	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

// RegisterRoutes wires all HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, sessions *store.SessionStore, games *store.GameStore) {
	sh := newSessionHandler(sessions)
	gh := newGameHandler(sessions, games)
	jh := newJoinHandler(sessions, games)
	mh := newMoveHandler(sessions, games)
	rh := newResignHandler(sessions, games)
	dh := newDrawHandler(sessions, games)

	// Sessions
	mux.HandleFunc("POST /sessions", sh.create)

	// Games lobby
	mux.HandleFunc("GET /games", requireSession(sessions, gh.list))
	mux.HandleFunc("POST /games", requireSession(sessions, gh.create))
	mux.HandleFunc("GET /games/{id}", requireSession(sessions, gh.detail))
	mux.HandleFunc("DELETE /games/{id}", requireSession(sessions, gh.cancel))

	// Game actions
	mux.HandleFunc("POST /games/{id}/join", requireSession(sessions, jh.join))
	mux.HandleFunc("POST /games/{id}/moves", requireSession(sessions, mh.move))
	mux.HandleFunc("POST /games/{id}/resign", requireSession(sessions, rh.resign))
	mux.HandleFunc("POST /games/{id}/draw", requireSession(sessions, dh.draw))
}
