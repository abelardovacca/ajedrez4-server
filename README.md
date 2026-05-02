# ajedrez4-server

HTTP REST chess game server backed by [ajedrez4-common](../ajedrez4-common).

## Running

```bash
go build ./cmd/server && ./server
# Listens on :8080 by default

PORT=9090 ./server  # custom port
```

Or without building:

```bash
go run ./cmd/server
```

## Authentication

All endpoints (except `POST /sessions`) require the `X-Session-ID` header obtained from session creation.

## Quick Start (curl)

### Create sessions

```bash
curl -s -X POST http://localhost:8080/sessions \
  -H "Content-Type: application/json" \
  -d '{"display_name": "Alice"}'
# → {"session_id": "SESSION_A", "display_name": "Alice", "created_at": "..."}

curl -s -X POST http://localhost:8080/sessions \
  -H "Content-Type: application/json" \
  -d '{"display_name": "Bob"}'
# → {"session_id": "SESSION_B", ...}
```

### Create and join a game

```bash
# Alice creates a game (she is white)
curl -s -X POST http://localhost:8080/games \
  -H "X-Session-ID: SESSION_A"
# → {"game_id": "GAME_ID", "status": "open", "white_player": "Alice", ...}

# Bob joins (he is black)
curl -s -X POST http://localhost:8080/games/GAME_ID/join \
  -H "X-Session-ID: SESSION_B"
# → {"game_id": "GAME_ID", "status": "active", ...}
```

### Play moves

```bash
curl -s -X POST http://localhost:8080/games/GAME_ID/moves \
  -H "X-Session-ID: SESSION_A" \
  -H "Content-Type: application/json" \
  -d '{"notation": "e2e4"}'

curl -s -X POST http://localhost:8080/games/GAME_ID/moves \
  -H "X-Session-ID: SESSION_B" \
  -H "Content-Type: application/json" \
  -d '{"notation": "e7e5"}'
```

### View game state

```bash
curl -s http://localhost:8080/games/GAME_ID \
  -H "X-Session-ID: SESSION_A"
# → {"status": "active", "fen": "...", "move_history": ["e2e4", "e7e5"], ...}
```

### Resign

```bash
curl -s -X POST http://localhost:8080/games/GAME_ID/resign \
  -H "X-Session-ID: SESSION_B"
# → {"status": "finished", "outcome": {"result": "white_wins", "winner": "white"}}
```

### Claim a draw

```bash
# Valid claim types: "threefold_repetition", "fifty_move_rule"
curl -s -X POST http://localhost:8080/games/GAME_ID/draw \
  -H "X-Session-ID: SESSION_A" \
  -H "Content-Type: application/json" \
  -d '{"claim_type": "fifty_move_rule"}'
```

### Cancel an open game

```bash
# Only the creator (white) can cancel an open game
curl -s -X DELETE http://localhost:8080/games/GAME_ID \
  -H "X-Session-ID: SESSION_A"
# → 204 No Content
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /sessions | Create a session |
| GET | /games | List all games |
| POST | /games | Create a game |
| GET | /games/{id} | Get game detail |
| DELETE | /games/{id} | Cancel open game |
| POST | /games/{id}/join | Join an open game |
| POST | /games/{id}/moves | Submit a move |
| POST | /games/{id}/resign | Resign |
| POST | /games/{id}/draw | Claim a draw |

## Move Notation

Moves use UCI coordinate notation: `e2e4`, `g1f3`, `e7e8q` (pawn promotion to queen).

## Development

```bash
go test ./...           # run all tests
go test -race ./...     # run with race detector
go vet ./...            # static analysis
```
