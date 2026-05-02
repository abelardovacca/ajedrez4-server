package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"
	"encoding/json"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/handler"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

func newIntegrationServer(t *testing.T) *httptest.Server {
	t.Helper()
	sessions := store.NewSessionStore()
	games := store.NewGameStore()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, sessions, games)
	return httptest.NewServer(mux)
}

func postJSONInt(t *testing.T, srv *httptest.Server, path string, body any, sessionID string) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	return resp
}

func decodeJSONInt(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// TestFullGame_FoolsMate plays through Fool's Mate:
// 1. f2f3  2. e7e5  3. g2g4  4. d8h4#  → black wins
func TestFullGame_FoolsMate(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	// Create sessions
	var white, black api.SessionResponse
	decodeJSONInt(t, postJSONInt(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "White"}, ""), &white)
	decodeJSONInt(t, postJSONInt(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Black"}, ""), &black)

	// Create game
	var game api.GameSummaryResponse
	decodeJSONInt(t, postJSONInt(t, srv, "/games", map[string]any{}, white.SessionID), &game)

	// Join game
	joinResp := postJSONInt(t, srv, fmt.Sprintf("/games/%s/join", game.GameID), map[string]any{}, black.SessionID)
	if joinResp.StatusCode != http.StatusOK {
		t.Fatalf("join: expected 200, got %d", joinResp.StatusCode)
	}

	// Fool's Mate sequence
	moves := []struct {
		notation  string
		sessionID string
	}{
		{"f2f3", white.SessionID},
		{"e7e5", black.SessionID},
		{"g2g4", white.SessionID},
		{"d8h4", black.SessionID}, // checkmate
	}

	var lastMoveResp api.MoveResponse
	for i, m := range moves {
		resp := postJSONInt(t, srv, fmt.Sprintf("/games/%s/moves", game.GameID),
			api.SubmitMoveRequest{Notation: m.notation}, m.sessionID)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("move %d (%s): expected 200, got %d", i+1, m.notation, resp.StatusCode)
		}
		decodeJSONInt(t, resp, &lastMoveResp)
	}

	if lastMoveResp.Status != "finished" {
		t.Errorf("expected finished after checkmate, got %q", lastMoveResp.Status)
	}
	if lastMoveResp.Outcome == nil {
		t.Fatal("expected outcome after checkmate")
	}
	if lastMoveResp.Outcome.Result != "black_wins" {
		t.Errorf("expected black_wins, got %q", lastMoveResp.Outcome.Result)
	}

	// Verify GET /games/{id} reflects finished state
	req, _ := http.NewRequest(http.MethodGet, srv.URL+fmt.Sprintf("/games/%s", game.GameID), nil)
	req.Header.Set("X-Session-ID", white.SessionID)
	detailResp, _ := http.DefaultClient.Do(req)
	var detail api.GameDetailResponse
	decodeJSONInt(t, detailResp, &detail)
	if detail.Status != "finished" {
		t.Errorf("GET game detail: expected finished, got %q", detail.Status)
	}
	if len(detail.MoveHistory) != 4 {
		t.Errorf("expected 4 moves in history, got %d", len(detail.MoveHistory))
	}
}
