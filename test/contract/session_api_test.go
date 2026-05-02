package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
	"github.com/ajedrez4/ajedrez4-server/internal/handler"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	sessions := store.NewSessionStore()
	games := store.NewGameStore()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, sessions, games)
	return httptest.NewServer(mux)
}

func postJSON(t *testing.T, srv *httptest.Server, path string, body any, sessionID string) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func TestCreateSession_ValidName(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var body api.SessionResponse
	decodeJSON(t, resp, &body)
	if body.SessionID == "" {
		t.Error("expected non-empty session_id")
	}
	if body.DisplayName != "Alice" {
		t.Errorf("expected display_name=Alice, got %q", body.DisplayName)
	}
}

func TestCreateSession_UniqueIDs(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	var r1, r2 api.SessionResponse
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &r1)
	decodeJSON(t, postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Alice"}, ""), &r2)
	if r1.SessionID == r2.SessionID {
		t.Errorf("expected distinct IDs, got both %q", r1.SessionID)
	}
}

func TestCreateSession_EmptyName(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: ""}, "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body api.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Error.Code != "invalid_display_name" {
		t.Errorf("expected invalid_display_name, got %q", body.Error.Code)
	}
}

func TestCreateSession_NameTooLong(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	long := strings.Repeat("a", 33)
	resp := postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: long}, "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateSession_NonPrintableChars(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "abc\x01def"}, "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateSession_MaxLengthName(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	name := strings.Repeat("a", 32)
	resp := postJSON(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: name}, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for 32-char name, got %d", resp.StatusCode)
	}
}
