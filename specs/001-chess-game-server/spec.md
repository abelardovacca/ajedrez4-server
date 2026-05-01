# Feature Specification: Chess Game Server

**Feature Branch**: `001-chess-game-server`  
**Created**: 2026-05-01  
**Status**: Draft  
**Input**: User description: "in ajedrez4-server, create a server for a game of chess game, that uses ajedrez4-common for the game itself. The server should allow to users to list, create and join games and play against each other. (for now, no authentication, any user with access to the server should be able to create a user session)"

## Clarifications

### Session 2026-05-01

- Q: What communication protocol should the server expose? → A: HTTP REST with JSON payloads.

## Clarifications

### Session 2026-05-01

- Q: What communication protocol should the server expose? → A: HTTP REST with JSON payloads.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create a User Session (Priority: P1)

As an application user, I can create a named session so that I have an identity within the server that lets me participate in games.

**Why this priority**: Every other interaction requires an identity. A session is the minimal prerequisite for all gameplay. Without it, no other feature can function independently.

**Independent Test**: Can be fully tested by sending a session creation request with a display name and verifying a unique session ID is returned, without needing games or move submission.

**Acceptance Scenarios**:

1. **Given** a user with no existing session, **When** they create a session with a display name, **Then** the server returns a unique session ID they can use for subsequent requests.
2. **Given** a user creates two sessions with the same display name, **When** both requests succeed, **Then** each session receives a distinct session ID.
3. **Given** a session creation request with an empty or invalid display name, **When** the server processes it, **Then** the server rejects it with a clear validation error.

   Display name rules: 1–32 printable characters, non-empty.

---

### User Story 2 - Create and Join Games (Priority: P2)

As a user with an active session, I can create a new game and wait for an opponent, or join an existing open game, so that two players can be matched for a chess match.

**Why this priority**: Matchmaking is the second prerequisite to gameplay. Once sessions exist, this enables the matchmaking flow that creates the match context required for move submission.

**Independent Test**: Can be fully tested in isolation by having two sessions — one creating a game and one joining it — and verifying both are assigned as players without submitting any moves.

**Acceptance Scenarios**:

1. **Given** an authenticated session, **When** the user creates a new game, **Then** a game is created in an open state assigned to that user as the creator, waiting for an opponent.
2. **Given** an open game and a second session, **When** that second user joins the game, **Then** the game transitions to an active state with both players assigned and play can begin.
3. **Given** an already-active game, **When** a third user attempts to join, **Then** the server rejects the join with a clear error.
4. **Given** a game creator, **When** they attempt to join their own game as the second player, **Then** the server rejects the join with a clear error.
5. **Given** an open game with no second player, **When** the creator cancels the game, **Then** the game is removed from the lobby and no longer joinable.

---

### User Story 3 - Play a Game (Priority: P3)

As a player in an active chess game, I can submit moves and view the current game state so that two players can complete a full game of chess against each other.

**Why this priority**: This is the core product value delivered once a match is established. It depends on sessions (P1) and game creation/joining (P2).

**Independent Test**: Can be fully tested independently by restoring a game from a FEN position and simulating alternating moves between two sessions until a terminal state is reached.

**Acceptance Scenarios**:

1. **Given** an active game and the side-to-move player's session, **When** a legal move is submitted, **Then** the server accepts the move, updates the game state, and returns the new state.
2. **Given** an active game, **When** the player who is NOT the side-to-move submits a move, **Then** the server rejects the move without changing game state.
3. **Given** an active game, **When** a move results in checkmate or stalemate, **Then** the game transitions to a finished state with the correct outcome.
4. **Given** an active game, **When** the active player resigns, **Then** the game ends immediately with the opponent declared the winner.
5. **Given** a game with eligible draw conditions, **When** the eligible player claims a draw, **Then** the game ends as a draw if the claim is valid, or is rejected with a reason if not eligible.
6. **Given** a finished game, **When** any player attempts to submit a move, **Then** the server rejects the move with a game-finished error.

---

### User Story 4 - List Available Games (Priority: P4)

As a user with an active session, I can browse open and active games so that I can find a game to join or follow.

**Why this priority**: Lobby browsing is a convenience feature. It adds discoverability but does not block game creation or play.

**Independent Test**: Can be fully tested in isolation by creating several games in various states and verifying the listing returns accurate counts and status per game.

**Acceptance Scenarios**:

1. **Given** several games in open, active, and finished states, **When** a user requests the game list, **Then** the server returns all games (open, active, and finished) with their current status, player names, and creation time.
2. **Given** no games exist, **When** a user requests the game list, **Then** the server returns an empty list without error.
3. **Given** a specific game ID, **When** a user requests that game's detail, **Then** the server returns the full game state including the current board position.

---

### Edge Cases

- Submitting a move with an invalid session ID.
- Submitting a move in a game the session is not part of.
- Requesting a game that does not exist.
- Joining a game that has already been joined by another player after listing but before joining (race condition: first request wins; the loser receives a conflict error).
- Submitting a move using invalid UCI notation.
- Session IDs expiring or being reused (not in scope for v1 but must not cause silent data corruption).
- A player abandoning a game without resigning (game remains in active state indefinitely in v1).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow any caller to create a user session identified by a display name (1–32 printable characters, non-empty), returning a unique session token for use in subsequent requests.
- **FR-002**: System MUST allow users with a valid session to create a new chess game, placing it in an open state awaiting a second player.
- **FR-003**: System MUST allow users with a valid session to join an open game as the second player, transitioning the game to active and assigning player colors.
- **FR-004**: System MUST prevent a game creator from joining their own game as the second player.
- **FR-005**: System MUST prevent more than two players from joining a single game.
- **FR-019**: System MUST allow the creator of an open (un-joined) game to cancel it, removing it from the lobby; cancelling an active or finished game is not permitted.
- **FR-006**: System MUST allow players to submit moves in UCI coordinate notation; move legality is validated and enforced by ajedrez4-common.
- **FR-007**: System MUST reject moves submitted by a player whose session is not the active side-to-move in the game.
- **FR-008**: System MUST reject moves submitted by sessions that are not participants in the game.
- **FR-009**: System MUST allow the active side-to-move player to resign, immediately ending the game with the opponent as winner.
- **FR-010**: System MUST allow the active side-to-move player to claim a draw by threefold repetition or 50-move rule; the claim is validated and resolved by ajedrez4-common.
- **FR-011**: System MUST expose the current game state including board layout (FEN), active player, move history, and game status on demand.
- **FR-012**: System MUST return the list of all games (open, active, and finished) with their status, participant display names, and creation time.
- **FR-013**: System MUST return a specific game's full detail including current board state.
- **FR-014**: System MUST use ajedrez4-common as the sole source of chess rule enforcement; no game logic is reimplemented in the server.
- **FR-015**: System MUST handle concurrent requests to join the same open game safely; the first request wins, the game transitions to active, and all subsequent join attempts receive a conflict error.
- **FR-016**: System MUST return structured JSON error responses with human-readable messages and machine-readable error codes for all failure conditions.
- **FR-017**: System MUST NOT require authentication; any caller with network access may create a session and participate in games.
- **FR-018**: System MUST expose its functionality via an HTTP REST API with JSON request and response payloads.

### Key Entities

- **UserSession**: An ephemeral identity with a unique session ID and a user-provided display name (1–32 printable characters). Not persisted across server restarts.
- **Game**: A chess match instance with two player slots, game status (open/active/finished), color assignments, and a reference to the underlying ajedrez4-common game state.
- **Player**: A UserSession participating in a Game, assigned to either the white or black side.
- **GameListing**: A summary view of a game suitable for lobby display: game ID, status, player display names, creation time.
- **GameDetail**: Full view of a game including current board position (FEN), move history, active player, and outcome if finished.
- **MoveRequest**: A command from a player to advance the game: session token + game ID + UCI move notation.
- **DrawClaim**: A command from a player to claim a draw: session token + game ID + claim type (threefold or fifty-move).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Two players can complete a full game (session creation → game creation → join → play to checkmate or resignation) in a single acceptance test run with 100% success.
- **SC-002**: All illegal move submissions (wrong turn, illegal notation, non-participant) are rejected with correct error codes in 100% of contract test cases.
- **SC-003**: No chess rule logic is duplicated in the server; 100% of move validation, check detection, and end-game detection is delegated to ajedrez4-common.
- **SC-004**: Concurrent join attempts on the same open game result in exactly one success (first request wins) and all others receiving a conflict error with a machine-readable code.
- **SC-005**: All API error responses include a machine-readable error code and a human-readable message, verifiable across all error test scenarios.

## Assumptions

- ajedrez4-common is available as a local Go module dependency and is the sole source of chess rule enforcement.
- Game state is stored in-memory for v1; persistence across server restarts is out of scope.
- Client-to-server communication uses HTTP REST with JSON payloads in a request/response (polling) model for v1; server-push or real-time notifications are out of scope.
- Color assignment (white/black) is determined by the server at game-join time: creator is white, joiner is black.
- Session tokens are valid for the lifetime of the server process; no session expiration in v1.
- Only standard chess games are supported; variants are out of scope.
- No rate limiting, abuse prevention, or authorization beyond session identification is in scope for v1.
- A player who leaves a game without resigning leaves the game in active state indefinitely; no timeout or abandonment handling in v1.
- Any user may observe any game's state (no private games).
