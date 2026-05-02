package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// SessionResponse is returned by POST /sessions.
type SessionResponse struct {
	SessionID   string    `json:"session_id"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// GameSummaryResponse is used in GET /games and POST /games.
type GameSummaryResponse struct {
	GameID      string    `json:"game_id"`
	Status      string    `json:"status"`
	WhitePlayer string    `json:"white_player"`
	BlackPlayer *string   `json:"black_player"`
	CreatedAt   time.Time `json:"created_at"`
}

// OutcomeResponse represents the final result of a finished game.
type OutcomeResponse struct {
	Result     string  `json:"result"`
	Winner     *string `json:"winner"`
	DrawReason *string `json:"draw_reason"`
}

// GameDetailResponse is returned by GET /games/{id} and POST /games/{id}/join.
type GameDetailResponse struct {
	GameID      string           `json:"game_id"`
	Status      string           `json:"status"`
	WhitePlayer string           `json:"white_player"`
	BlackPlayer *string          `json:"black_player"`
	CreatedAt   time.Time        `json:"created_at"`
	ActiveColor *string          `json:"active_color"`
	FEN         *string          `json:"fen"`
	MoveHistory []string         `json:"move_history"`
	Outcome     *OutcomeResponse `json:"outcome"`
}

// MoveResponse is returned by POST /games/{id}/moves.
type MoveResponse struct {
	Accepted    bool             `json:"accepted"`
	GameID      string           `json:"game_id"`
	Status      string           `json:"status"`
	ActiveColor *string          `json:"active_color"`
	FEN         string           `json:"fen"`
	MoveHistory []string         `json:"move_history"`
	Outcome     *OutcomeResponse `json:"outcome"`
}

// ResignResponse is returned by POST /games/{id}/resign.
type ResignResponse struct {
	GameID  string          `json:"game_id"`
	Status  string          `json:"status"`
	Outcome OutcomeResponse `json:"outcome"`
}

// DrawResponse is returned by POST /games/{id}/draw.
type DrawResponse struct {
	Accepted         bool    `json:"accepted"`
	GameID           string  `json:"game_id"`
	Status           string  `json:"status"`
	RejectionReason  *string `json:"rejection_reason,omitempty"`
}

// GamesListResponse wraps the list of games for GET /games.
type GamesListResponse struct {
	Games []GameSummaryResponse `json:"games"`
}

// ErrorDetail carries the machine-readable code and human-readable message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the envelope for all error responses.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a structured error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{Error: ErrorDetail{Code: code, Message: message}})
}
