package contract

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
)

// setupActiveGame creates two sessions, a game, and joins it.
// Returns (whiteSessionID, blackSessionID, gameID).
func setupActiveGame(t *testing.T, srv *testServerHelper) (string, string, string) {
	t.Helper()
	var r1, r2 api.SessionResponse
	decodeJSON(t, postJSON(t, srv.s, "/sessions", api.CreateSessionRequest{DisplayName: "White"}, ""), &r1)
	decodeJSON(t, postJSON(t, srv.s, "/sessions", api.CreateSessionRequest{DisplayName: "Black"}, ""), &r2)
	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv.s, "/games", map[string]any{}, r1.SessionID), &game)
	postJSON(t, srv.s, fmt.Sprintf("/games/%s/join", game.GameID), map[string]any{}, r2.SessionID)
	return r1.SessionID, r2.SessionID, game.GameID
}

// testServerHelper wraps httptest.Server.
type testServerHelper struct {
	s *httptest.Server
}

func newHelper(t *testing.T) *testServerHelper {
	t.Helper()
	srv := newTestServer(t)
	t.Cleanup(srv.Close)
	return &testServerHelper{s: srv}
}

// moveReq posts a move to the given game.
func moveReq(t *testing.T, srv *httptest.Server, gameID, sessionID, notation string) *http.Response {
	t.Helper()
	return postJSON(t, srv, fmt.Sprintf("/games/%s/moves", gameID),
		api.SubmitMoveRequest{Notation: notation}, sessionID)
}

func TestMoveAPI_LegalMove(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	w, _, gid := setupActiveGame(t, &testServerHelper{srv})

	resp := moveReq(t, srv, gid, w, "e2e4")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body api.MoveResponse
	decodeJSON(t, resp, &body)
	if !body.Accepted {
		t.Error("expected accepted=true")
	}
	if body.FEN == "" {
		t.Error("expected non-empty FEN")
	}
	if len(body.MoveHistory) != 1 {
		t.Errorf("expected 1 move in history, got %d", len(body.MoveHistory))
	}
}

func TestMoveAPI_IllegalMove(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	w, _, gid := setupActiveGame(t, &testServerHelper{srv})

	resp := moveReq(t, srv, gid, w, "e2e5") // illegal jump
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "illegal_move" {
		t.Errorf("expected illegal_move, got %q", errResp.Error.Code)
	}
}

func TestMoveAPI_WrongTurn(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	_, b, gid := setupActiveGame(t, &testServerHelper{srv})

	// Black tries to move first (white's turn)
	resp := moveReq(t, srv, gid, b, "e7e5")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "wrong_turn" {
		t.Errorf("expected wrong_turn, got %q", errResp.Error.Code)
	}
}

func TestMoveAPI_NonParticipant(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	_, _, gid := setupActiveGame(t, &testServerHelper{srv})

	var outsider api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Outsider"}, ""), &outsider)

	resp := moveReq(t, srv, gid, outsider.SessionID, "e2e4")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "not_a_participant" {
		t.Errorf("expected not_a_participant, got %q", errResp.Error.Code)
	}
}

func TestMoveAPI_GameNotActive(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var sessResp api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &sessResp)
	var game api.GameSummaryResponse
	decodeJSON(t, postJSON(t, srv, "/games", map[string]any{}, sessResp.SessionID), &game)

	resp := moveReq(t, srv, game.GameID, sessResp.SessionID, "e2e4")
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestMoveAPI_InvalidNotation(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	w, _, gid := setupActiveGame(t, &testServerHelper{srv})

	resp := moveReq(t, srv, gid, w, "notmove")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "invalid_notation" {
		t.Errorf("expected invalid_notation, got %q", errResp.Error.Code)
	}
}

func TestResignAPI_ByParticipant(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	w, _, gid := setupActiveGame(t, &testServerHelper{srv})

	resp := postJSON(t, srv, fmt.Sprintf("/games/%s/resign", gid), map[string]any{}, w)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body api.ResignResponse
	decodeJSON(t, resp, &body)
	if body.Status != "finished" {
		t.Errorf("expected finished, got %q", body.Status)
	}
}

func TestResignAPI_ByNonParticipant(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	_, _, gid := setupActiveGame(t, &testServerHelper{srv})

	var outsider api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Outsider"}, ""), &outsider)

	resp := postJSON(t, srv, fmt.Sprintf("/games/%s/resign", gid), map[string]any{}, outsider.SessionID)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestDrawAPI_IneligibleClaim(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	w, _, gid := setupActiveGame(t, &testServerHelper{srv})

	// Fresh game — no draw conditions met
	resp := postJSON(t, srv, fmt.Sprintf("/games/%s/draw", gid),
		api.ClaimDrawRequest{ClaimType: "fifty_move_rule"}, w)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body api.DrawResponse
	decodeJSON(t, resp, &body)
	if body.Accepted {
		t.Error("expected accepted=false for ineligible draw")
	}
	if body.RejectionReason == nil {
		t.Error("expected rejection_reason")
	}
}

func TestDrawAPI_UnknownClaimType(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	w, _, gid := setupActiveGame(t, &testServerHelper{srv})

	resp := postJSON(t, srv, fmt.Sprintf("/games/%s/draw", gid),
		api.ClaimDrawRequest{ClaimType: "unknown_claim"}, w)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var errResp api.ErrorResponse
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "invalid_draw_claim_type" {
		t.Errorf("expected invalid_draw_claim_type, got %q", errResp.Error.Code)
	}
}
