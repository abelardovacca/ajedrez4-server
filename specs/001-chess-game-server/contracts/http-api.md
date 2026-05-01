# API Contract: Chess Game Server

**Feature**: `001-chess-game-server`  
**Protocol**: HTTP REST, JSON payloads  
**Base URL**: `http://host:port` (port configurable, default 8080)  
**Session Auth**: All endpoints except `POST /sessions` require the `X-Session-ID` header.

---

## Common Response Envelope

### Success

Individual endpoints return their own response bodies directly (not wrapped).

### Error

All error responses use this envelope:

```json
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

---

## Sessions

### POST /sessions

Create a new user session.

**Request body**:
```json
{
  "display_name": "Alice"
}
```

| Field | Type | Rules |
|-------|------|-------|
| `display_name` | string | 1–32 printable chars, required |

**Response `201 Created`**:
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "display_name": "Alice",
  "created_at": "2026-05-01T10:00:00Z"
}
```

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 400 | `invalid_display_name` | Empty, >32 chars, or non-printable chars |

---

## Games

### GET /games

List all games in all statuses.

**Headers**: `X-Session-ID: <session_id>`

**Response `200 OK`**:
```json
{
  "games": [
    {
      "game_id": "...",
      "status": "open",
      "white_player": "Alice",
      "black_player": null,
      "created_at": "2026-05-01T10:00:00Z"
    },
    {
      "game_id": "...",
      "status": "active",
      "white_player": "Alice",
      "black_player": "Bob",
      "created_at": "2026-05-01T10:01:00Z"
    },
    {
      "game_id": "...",
      "status": "finished",
      "white_player": "Alice",
      "black_player": "Bob",
      "created_at": "2026-05-01T09:00:00Z"
    }
  ]
}
```

`black_player` is `null` for open games. `games` is an empty array when no games exist.

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 401 | `missing_session_header` | `X-Session-ID` header absent |
| 401 | `session_not_found` | Session ID not in store |

---

### POST /games

Create a new game. Caller becomes the white player.

**Headers**: `X-Session-ID: <session_id>`

**Request body**: empty (`{}` or no body)

**Response `201 Created`**:
```json
{
  "game_id": "...",
  "status": "open",
  "white_player": "Alice",
  "black_player": null,
  "created_at": "2026-05-01T10:00:00Z"
}
```

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 401 | `missing_session_header` | Header absent |
| 401 | `session_not_found` | Session not found |

---

### GET /games/{id}

Get full game detail including current board position.

**Headers**: `X-Session-ID: <session_id>`

**Response `200 OK`**:
```json
{
  "game_id": "...",
  "status": "active",
  "white_player": "Alice",
  "black_player": "Bob",
  "created_at": "2026-05-01T10:01:00Z",
  "active_color": "white",
  "fen": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
  "move_history": ["e2e4"],
  "outcome": null
}
```

- `active_color`: `"white"` or `"black"` (null if game not active)
- `fen`: current position in FEN (null if game not yet started, i.e. open)
- `outcome`: null while in progress; when finished:
```json
"outcome": {
  "result": "white_wins",
  "winner": "white",
  "draw_reason": null
}
```
`result` values: `"white_wins"`, `"black_wins"`, `"draw"`, `"in_progress"`

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 401 | `missing_session_header` | Header absent |
| 401 | `session_not_found` | Session not found |
| 404 | `game_not_found` | Game ID not found |

---

### DELETE /games/{id}

Cancel an open game. Only the creator (white player) may cancel. Only `open` games can be cancelled.

**Headers**: `X-Session-ID: <session_id>`

**Response `204 No Content`**: (empty body)

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 401 | `missing_session_header` | Header absent |
| 401 | `session_not_found` | Session not found |
| 403 | `cannot_cancel_game` | Caller is not the creator, or game is not open |
| 404 | `game_not_found` | Game ID not found |

---

### POST /games/{id}/join

Join an open game as the black player.

**Headers**: `X-Session-ID: <session_id>`

**Request body**: empty (`{}` or no body)

**Response `200 OK`**:
```json
{
  "game_id": "...",
  "status": "active",
  "white_player": "Alice",
  "black_player": "Bob",
  "created_at": "2026-05-01T10:00:00Z",
  "active_color": "white",
  "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
  "move_history": [],
  "outcome": null
}
```

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 401 | `missing_session_header` | Header absent |
| 401 | `session_not_found` | Session not found |
| 403 | `cannot_join_own_game` | Caller is the creator |
| 404 | `game_not_found` | Game ID not found |
| 409 | `game_not_open` | Game is not in `open` status (already joined or finished) |

---

## Game Actions

### POST /games/{id}/moves

Submit a move in an active game.

**Headers**: `X-Session-ID: <session_id>`

**Request body**:
```json
{
  "notation": "e2e4"
}
```

**Response `200 OK`** (move accepted):
```json
{
  "accepted": true,
  "game_id": "...",
  "status": "active",
  "active_color": "black",
  "fen": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
  "move_history": ["e2e4"],
  "outcome": null
}
```

When the move ends the game (`status` becomes `"finished"`):
```json
{
  "accepted": true,
  "game_id": "...",
  "status": "finished",
  "active_color": null,
  "fen": "...",
  "move_history": ["f2f3", "e7e5", "g2g4", "d8h4"],
  "outcome": {
    "result": "black_wins",
    "winner": "black",
    "draw_reason": null
  }
}
```

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 400 | `invalid_notation` | UCI notation malformed |
| 401 | `missing_session_header` | Header absent |
| 401 | `session_not_found` | Session not found |
| 403 | `not_a_participant` | Session is not a player in this game |
| 403 | `wrong_turn` | Not this player's turn |
| 404 | `game_not_found` | Game not found |
| 409 | `game_not_active` | Game is not in `active` status |
| 422 | `illegal_move` | Move is illegal in current position |

---

### POST /games/{id}/resign

Resign from an active game.

**Headers**: `X-Session-ID: <session_id>`

**Request body**: empty

**Response `200 OK`**:
```json
{
  "game_id": "...",
  "status": "finished",
  "outcome": {
    "result": "white_wins",
    "winner": "white",
    "draw_reason": null
  }
}
```

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 401 | `missing_session_header` | Header absent |
| 401 | `session_not_found` | Session not found |
| 403 | `not_a_participant` | Session is not a player |
| 404 | `game_not_found` | Game not found |
| 409 | `game_not_active` | Game is not active |

---

### POST /games/{id}/draw

Claim a draw by threefold repetition or 50-move rule.

**Headers**: `X-Session-ID: <session_id>`

**Request body**:
```json
{
  "claim_type": "threefold_repetition"
}
```

`claim_type` values: `"threefold_repetition"`, `"fifty_move_rule"`

**Response `200 OK`** (claim accepted — draw granted):
```json
{
  "accepted": true,
  "game_id": "...",
  "status": "finished",
  "outcome": {
    "result": "draw",
    "winner": null,
    "draw_reason": "threefold_repetition"
  }
}
```

**Response `200 OK`** (claim rejected — game continues):
```json
{
  "accepted": false,
  "rejection_reason": {
    "code": "draw_claim_ineligible",
    "message": "Threefold repetition condition has not been met"
  },
  "game_id": "...",
  "status": "active"
}
```

**Errors**:
| Status | Code | Condition |
|--------|------|-----------|
| 400 | `invalid_draw_claim_type` | Unknown claim type string |
| 401 | `missing_session_header` | Header absent |
| 401 | `session_not_found` | Session not found |
| 403 | `not_a_participant` | Session is not a player |
| 403 | `wrong_turn` | Only the side-to-move may claim a draw |
| 404 | `game_not_found` | Game not found |
| 409 | `game_not_active` | Game is not active |
