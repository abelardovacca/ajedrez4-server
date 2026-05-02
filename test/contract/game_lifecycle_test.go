package contract

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
)

func TestGameLifecycle_CreateGame(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var sessResp api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &sessResp)
	sid := sessResp.SessionID

	resp := postJSON(t, srv, "/games", map[string]any{}, sid)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var g api.GameSummaryResponse
	decodeJSON(t, resp, &g)
	if g.Status != "open" {
		t.Errorf("expected open, got %q", g.Status)
	}
	if g.GameID == "" {
		t.Error("expected game_id")
	}
	if g.BlackPlayer != nil {
		t.Error("expected null black_player for open game")
	}
}

func TestGameLifecycle_JoinGame(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var r1, r2 api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &r1)
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Bob"}, ""), &r2)

	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, r1.SessionID), &game)

	joinResp := postJSON(t, srv, fmt.Sprintf("/games/%s/join", game.GameID), map[string]any{}, r2.SessionID)
	if joinResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", joinResp.StatusCode)
	}
	var detail api.GameDetailResponse
	decodeJSON(t, joinResp, &detail)
	if detail.Status != "active" {
		t.Errorf("expected active, got %q", detail.Status)
	}
	if detail.FEN == nil || *detail.FEN == "" {
		t.Error("expected FEN after join")
	}
}

func TestGameLifecycle_JoinAlreadyActive(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var r1, r2, r3 api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &r1)
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Bob"}, ""), &r2)
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Carol"}, ""), &r3)

	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, r1.SessionID), &game)
	postJSON(t, srv, fmt.Sprintf("/games/%s/join", game.GameID), map[string]any{}, r2.SessionID)

	resp := postJSON(t, srv, fmt.Sprintf("/games/%s/join", game.GameID), map[string]any{}, r3.SessionID)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "game_not_open" {
		t.Errorf("expected game_not_open, got %q", errResp.Error.Code)
	}
}

func TestGameLifecycle_JoinOwnGame(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var sessResp api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &sessResp)

	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, sessResp.SessionID), &game)

	resp := postJSON(t, srv, fmt.Sprintf("/games/%s/join", game.GameID), map[string]any{}, sessResp.SessionID)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "cannot_join_own_game" {
		t.Errorf("expected cannot_join_own_game, got %q", errResp.Error.Code)
	}
}

func TestGameLifecycle_CancelByCreator(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var sessResp api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &sessResp)

	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, sessResp.SessionID), &game)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+fmt.Sprintf("/games/%s", game.GameID), nil)
	req.Header.Set("X-Session-ID", sessResp.SessionID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestGameLifecycle_CancelByNonCreator(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var r1, r2 api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &r1)
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Bob"}, ""), &r2)

	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, r1.SessionID), &game)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+fmt.Sprintf("/games/%s", game.GameID), nil)
	req.Header.Set("X-Session-ID", r2.SessionID)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "cannot_cancel_game" {
		t.Errorf("expected cannot_cancel_game, got %q", errResp.Error.Code)
	}
}

func TestGameLifecycle_MissingSessionHeader(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postJSON(t, srv, "/games", map[string]any{}, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "missing_session_header" {
		t.Errorf("expected missing_session_header, got %q", errResp.Error.Code)
	}
}

// --- GET /games and GET /games/{id} ---

func TestListGames_Empty(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var sessResp api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &sessResp)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/games", nil)
	req.Header.Set("X-Session-ID", sessResp.SessionID)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body api.GamesListResponse
	decodeJSON(t, resp, &body)
	if len(body.Games) != 0 {
		t.Errorf("expected empty list, got %d games", len(body.Games))
	}
}

func TestListGames_AllStatuses(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var r1, r2 api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &r1)
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Bob"}, ""), &r2)

	// Create open game
	var openGame api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, r1.SessionID), &openGame)

	// Create and join → active
	var g2 api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, r1.SessionID), &g2)
	postJSON(t, srv, fmt.Sprintf("/games/%s/join", g2.GameID), map[string]any{}, r2.SessionID)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/games", nil)
	req.Header.Set("X-Session-ID", r1.SessionID)
	resp, _ := http.DefaultClient.Do(req)
	var list api.GamesListResponse
	decodeJSON(t, resp, &list)
	if len(list.Games) != 2 {
		t.Errorf("expected 2 games, got %d", len(list.Games))
	}
}

func TestGetGameDetail_NotFound(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var sessResp api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &sessResp)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/games/nonexistent", nil)
	req.Header.Set("X-Session-ID", sessResp.SessionID)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "game_not_found" {
		t.Errorf("expected game_not_found, got %q", errResp.Error.Code)
	}
}

func TestGetGameDetail_OpenGameNullFEN(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var sessResp api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &sessResp)

	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, sessResp.SessionID), &game)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+fmt.Sprintf("/games/%s", game.GameID), nil)
	req.Header.Set("X-Session-ID", sessResp.SessionID)
	resp, _ := http.DefaultClient.Do(req)
	var detail api.GameDetailResponse
	decodeJSON(t, resp, &detail)
	if detail.FEN != nil {
		t.Errorf("expected null FEN for open game, got %q", *detail.FEN)
	}
}

func TestGetGameDetail_ActiveGameHasFEN(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var r1, r2 api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &r1)
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Bob"}, ""), &r2)

	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, r1.SessionID), &game)
	postJSON(t, srv, fmt.Sprintf("/games/%s/join", game.GameID), map[string]any{}, r2.SessionID)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+fmt.Sprintf("/games/%s", game.GameID), nil)
	req.Header.Set("X-Session-ID", r1.SessionID)
	resp, _ := http.DefaultClient.Do(req)
	var detail api.GameDetailResponse
	decodeJSON(t, resp, &detail)
	if detail.FEN == nil || *detail.FEN == "" {
		t.Error("expected non-null FEN for active game")
	}
	if detail.ActiveColor == nil {
		t.Error("expected active_color for active game")
	}
}
