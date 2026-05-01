# Quickstart: Chess Game Server

**Feature**: `001-chess-game-server`  
**Date**: 2026-05-01

## Running the Server

```bash
go run ./cmd/server
# Server listens on :8080 by default
```

## Playing a Full Game (curl)

### Step 1 — Create two player sessions

```bash
# Player 1 (will be white)
curl -s -X POST http://localhost:8080/sessions \
  -H "Content-Type: application/json" \
  -d '{"display_name": "Alice"}'
# → {"session_id": "SESSION_A", "display_name": "Alice", "created_at": "..."}

# Player 2 (will be black)
curl -s -X POST http://localhost:8080/sessions \
  -H "Content-Type: application/json" \
  -d '{"display_name": "Bob"}'
# → {"session_id": "SESSION_B", "display_name": "Bob", "created_at": "..."}
```

### Step 2 — Alice creates a game

```bash
curl -s -X POST http://localhost:8080/games \
  -H "X-Session-ID: SESSION_A"
# → {"game_id": "GAME_ID", "status": "open", "white_player": "Alice", ...}
```

### Step 3 — Bob joins the game

```bash
curl -s -X POST http://localhost:8080/games/GAME_ID/join \
  -H "X-Session-ID: SESSION_B"
# → {"game_id": "GAME_ID", "status": "active", "white_player": "Alice", "black_player": "Bob", ...}
```

### Step 4 — Play moves (alternating)

```bash
# Alice plays e2e4
curl -s -X POST http://localhost:8080/games/GAME_ID/moves \
  -H "X-Session-ID: SESSION_A" \
  -H "Content-Type: application/json" \
  -d '{"notation": "e2e4"}'

# Bob plays e7e5
curl -s -X POST http://localhost:8080/games/GAME_ID/moves \
  -H "X-Session-ID: SESSION_B" \
  -H "Content-Type: application/json" \
  -d '{"notation": "e7e5"}'
```

### Step 5 — View game state

```bash
curl -s http://localhost:8080/games/GAME_ID \
  -H "X-Session-ID: SESSION_A"
# → {"game_id": "...", "status": "active", "fen": "...", "move_history": ["e2e4", "e7e5"], ...}
```

### Step 6 — Resign

```bash
curl -s -X POST http://localhost:8080/games/GAME_ID/resign \
  -H "X-Session-ID: SESSION_B"
# → {"status": "finished", "outcome": {"result": "white_wins", "winner": "white"}}
```

---

## List All Games

```bash
curl -s http://localhost:8080/games \
  -H "X-Session-ID: SESSION_A"
# → {"games": [...]}
```

---

## Claim a Draw

```bash
# Only valid when the 50-move or threefold condition is met
curl -s -X POST http://localhost:8080/games/GAME_ID/draw \
  -H "X-Session-ID: SESSION_A" \
  -H "Content-Type: application/json" \
  -d '{"claim_type": "fifty_move_rule"}'
# → {"accepted": true, "status": "finished", "outcome": {"result": "draw", "draw_reason": "fifty_move_rule"}}
```

---

## Error Handling

All errors return:
```json
{
  "error": {
    "code": "wrong_turn",
    "message": "It is not your turn to move"
  }
}
```

Common HTTP status codes: `400` (bad input), `401` (invalid session), `403` (forbidden action), `404` (not found), `409` (conflict), `422` (illegal move).
