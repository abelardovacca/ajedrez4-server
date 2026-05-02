package handler

import (
	"context"
	"net/http"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/domain"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

type contextKey string

const sessionContextKey contextKey = "session"

// sessionFromContext retrieves the validated session from a request context.
func sessionFromContext(ctx context.Context) *domain.Session {
	sess, _ := ctx.Value(sessionContextKey).(*domain.Session)
	return sess
}

// requireSession is middleware that validates X-Session-ID and injects the
// session into the request context. Returns 401 on missing or unknown IDs.
func requireSession(sessions *store.SessionStore, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Session-ID")
		if id == "" {
			api.WriteError(w, http.StatusUnauthorized, domain.ErrMissingSessionHeader,
				"X-Session-ID header is required")
			return
		}
		sess, ok := sessions.GetByID(id)
		if !ok {
			api.WriteError(w, http.StatusUnauthorized, domain.ErrSessionNotFound,
				"Session not found")
			return
		}
		ctx := context.WithValue(r.Context(), sessionContextKey, sess)
		next(w, r.WithContext(ctx))
	}
}
