# Data Model: Chess Game Server

**Feature**: `001-chess-game-server`  
**Phase**: 1 — Design  
**Date**: 2026-05-01

## Entities

### Session

Represents an ephemeral user identity for the lifetime of the server process.

| Field | Type | Notes |
|-------|------|-------|
| `ID` | `string` (UUID v4) | Unique identifier; returned to caller on creation; used as `X-Session-ID` header value |
| `DisplayName` | `string` | 1–32 printable characters; provided by caller |
| `CreatedAt` | `time.Time` | Server-assigned creation timestamp |

**Validation rules**:
- `DisplayName` must be non-empty, 1–32 characters, all printable (no control characters)
- `ID` is server-generated and not caller-supplied

---

### Game

Represents a chess match lobby and its associated game state.

| Field | Type | Notes |
|-------|------|-------|
| `ID` | `string` (UUID v4) | Unique identifier |
| `Status` | `GameStatus` | `open` → `active` → `finished` (or `open` → `cancelled`) |
| `WhiteSessionID` | `string` | Session ID of the creator (white); assigned at creation |
| `BlackSessionID` | `string` | Session ID of the joiner (black); assigned at join; empty if open |
| `ChessGame` | `*chess.Game` | Pointer to the `ajedrez4-common` game state; nil if open |
| `CreatedAt` | `time.Time` | Server-assigned creation timestamp |

**Status transitions**:

```
open ──join──► active ──checkmate/stalemate/resign/draw──► finished
open ──cancel(creator)──► [removed from store]
```

**Invariants**:
- `WhiteSessionID` is always set when a Game exists
- `BlackSessionID` is set only in `active` or `finished` states
- `ChessGame` is non-nil only in `active` or `finished` states
- Only the creator (WhiteSessionID) may cancel; only when `open`
- Only two distinct sessions may be participants

---

### GameStatus

```
open      — created, waiting for second player
active    — both players assigned, game in progress
finished  — terminal state (checkmate, stalemate, draw, resignation)
```

---

### PlayerColor

```
white — assigned to the game creator
black — assigned to the joiner
```

---

### MoveCommand

Input for move submission (not persisted, used as a request DTO).

| Field | Type | Notes |
|-------|------|-------|
| `Notation` | `string` | UCI coordinate notation (e.g., `e2e4`, `e7e8q`) |

---

### DrawClaimCommand

Input for draw claim submission.

| Field | Type | Notes |
|-------|------|-------|
| `ClaimType` | `string` | `"threefold_repetition"` or `"fifty_move_rule"` |

---

## Aggregates

### Session Store

Holds all active sessions. Keyed by `Session.ID`.

- **Read**: by ID (lookup by `X-Session-ID` header on every request)
- **Write**: append-only (session creation); no deletion in v1

### Game Store

Holds all games in all states. Keyed by `Game.ID`.

- **Read**: full scan (list all), by ID (detail + operations)
- **Write**: create, join (status transition), cancel (removal), move/resign/draw (chess state update)
- **Concurrency**: join operation uses write-lock compare-and-swap: read `Status == open` and `BlackSessionID == ""` then assign — under the same write lock — ensuring first-request wins

---

## State Transition: Join (Critical Section)

```
acquire write lock
  if game.Status != "open" → return 409 Conflict
  if game.BlackSessionID != "" → return 409 Conflict  (defensive; covered by status check)
  if game.WhiteSessionID == joinerSessionID → return 403 Forbidden
  game.BlackSessionID = joinerSessionID
  game.Status = "active"
  game.ChessGame = chess.NewGame(game.ID)
release write lock
→ return 200 OK
```

---

## Error Codes

| Code | HTTP | Meaning |
|------|------|---------|
| `session_not_found` | 401 | `X-Session-ID` header missing or not in store |
| `missing_session_header` | 401 | `X-Session-ID` header absent |
| `invalid_display_name` | 400 | Display name fails length/character rules |
| `game_not_found` | 404 | Game ID does not exist |
| `game_not_open` | 409 | Join attempted on non-open game |
| `cannot_join_own_game` | 403 | Creator attempting to join their own game |
| `not_a_participant` | 403 | Session is not a player in this game |
| `wrong_turn` | 403 | Move submitted by the player not to move |
| `game_not_active` | 409 | Move/resign/draw on a non-active game |
| `illegal_move` | 422 | Move notation valid but illegal in position |
| `invalid_notation` | 400 | UCI notation malformed |
| `draw_claim_ineligible` | 422 | Draw condition not met |
| `invalid_draw_claim_type` | 400 | Unknown claim type string |
| `cannot_cancel_game` | 403 | Cancel attempted by non-creator or on non-open game |
