package integration

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ajedrez4/ajedrez4-server/internal/api"
)

// TestConcurrentJoin verifies first-request-wins semantics:
// two goroutines simultaneously join the same open game;
// exactly one should succeed (200) and one should get conflict (409).
func TestConcurrentJoin(t *testing.T) {
	srv := newIntegrationServer(t)
	defer srv.Close()

	// Create three sessions: one creator, two joiners
	var creator, joiner1, joiner2 api.SessionResponse
	decodeJSONInt(t, postJSONInt(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Creator"}, ""), &creator)
	decodeJSONInt(t, postJSONInt(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Joiner1"}, ""), &joiner1)
	decodeJSONInt(t, postJSONInt(t, srv, "/sessions", api.CreateSessionRequest{DisplayName: "Joiner2"}, ""), &joiner2)

	var game api.GameSummaryResponse
	decodeJSONInt(t, postJSONInt(t, srv, "/games", map[string]any{}, creator.SessionID), &game)

	var wg sync.WaitGroup
	results := make([]int, 2)

	for i, sid := range []string{joiner1.SessionID, joiner2.SessionID} {
		wg.Add(1)
		go func(idx int, sessionID string) {
			defer wg.Done()
			resp := postJSONInt(t, srv, fmt.Sprintf("/games/%s/join", game.GameID),
				map[string]any{}, sessionID)
			results[idx] = resp.StatusCode
		}(i, sid)
	}
	wg.Wait()

	successCount := 0
	conflictCount := 0
	for _, code := range results {
		switch code {
		case http.StatusOK:
			successCount++
		case http.StatusConflict:
			conflictCount++
		}
	}

	if successCount != 1 {
		t.Errorf("expected exactly 1 success, got %d (results: %v)", successCount, results)
	}
	if conflictCount != 1 {
		t.Errorf("expected exactly 1 conflict, got %d (results: %v)", conflictCount, results)
	}
}
