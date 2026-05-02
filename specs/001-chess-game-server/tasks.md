# Implementation Tasks: Chess Game Server

**Feature**: `001-chess-game-server`  
**Branch**: `001-chess-game-server`  
**Date**: 2026-05-01  
**Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)

## Summary

- **Total tasks**: 44
- **Parallelizable tasks**: 25
- **User stories covered**: 4 (US1–US4)
- **Suggested MVP scope**: Phase 1 + Phase 2 + Phase 3 (US1) — session creation only
- **MVP test**: `POST /sessions` creates a session, returns unique ID, rejects invalid names

---

## Phase 1: Setup

**Goal**: Initialize the Go module, project structure, and ajedrez4-common dependency.

- [X] T001 Initialize Go module `github.com/ajedrez4/ajedrez4-server` with `go mod init` in repository root
- [X] T002 Add `ajedrez4-common` dependency and local replace directive in `go.mod`: `require github.com/ajedrez4/ajedrez4-common v0.0.0` and `replace github.com/ajedrez4/ajedrez4-common => ../ajedrez4-common`
- [X] T003 Create directory structure: `cmd/server/`, `internal/domain/`, `internal/store/`, `internal/handler/`, `internal/api/`, `test/unit/`, `test/contract/`, `test/integration/`
- [X] T004 Create `cmd/server/main.go` stub (package main, empty `main()`)

---

## Phase 2: Foundational — Domain Types, API DTOs, Error Codes

**Goal**: All shared types, request/response DTOs, and error codes that all user story implementations depend on. Must be complete before any handler can be written.

- [X] T005 [P] Implement `Session` entity in `internal/domain/session.go`: struct with `ID string`, `DisplayName string`, `CreatedAt time.Time`; display name validation function (1–32 printable chars, non-empty)
- [X] T006 [P] Implement `Game` entity in `internal/domain/game.go`: struct with `ID string`, `Status GameStatus`, `WhiteSessionID string`, `BlackSessionID string`, `ChessGame *chess.Game`, `CreatedAt time.Time`; `GameStatus` type (`open`, `active`, `finished`)
- [X] T007 [P] Implement domain error codes in `internal/domain/errors.go`: constants for all error codes from data-model.md (`session_not_found`, `missing_session_header`, `invalid_display_name`, `game_not_found`, `game_not_open`, `cannot_join_own_game`, `not_a_participant`, `wrong_turn`, `game_not_active`, `illegal_move`, `invalid_notation`, `draw_claim_ineligible`, `invalid_draw_claim_type`, `cannot_cancel_game`)
- [X] T008 [P] Implement JSON request DTOs in `internal/api/requests.go`: `CreateSessionRequest{DisplayName string}`, `SubmitMoveRequest{Notation string}`, `ClaimDrawRequest{ClaimType string}`
- [X] T009 [P] Implement JSON response DTOs in `internal/api/responses.go`: `SessionResponse`, `GameSummaryResponse`, `GameDetailResponse`, `MoveResponse{FEN string, ActiveColor string, MoveHistory []string, Outcome *string}`, `ResignResponse{Outcome string}`, `DrawResponse{Accepted bool, RejectionReason string omitempty}`, `ErrorResponse{Error ErrorDetail}`, `ErrorDetail{Code string; Message string}`; helper `WriteError(w, status, code, message)`; helper `WriteJSON(w, status, v)`

---

## Phase 3: User Story 1 — Create a User Session (P1)

**Story goal**: Any caller can `POST /sessions` with a display name and receive a unique session ID.

**Independent test criteria**: `POST /sessions` with valid name → 201 + unique session ID; with empty name → 400 `invalid_display_name`; two requests with same name → two distinct IDs.

- [X] T010 [P] [US1] Implement `SessionStore` in `internal/store/session_store.go`: `sync.RWMutex`-protected map; methods `Create(session *domain.Session)` and `GetByID(id string) (*domain.Session, bool)`; UUID v4 generation via `crypto/rand`
- [X] T011 [US1] Implement `POST /sessions` handler in `internal/handler/session_handler.go`: parse `CreateSessionRequest`, validate display name, create session via store, return `SessionResponse` 201; return 400 `invalid_display_name` on validation failure
- [X] T012 [P] [US1] Unit test: `SessionStore` in `test/unit/session_store_test.go` — create returns unique IDs, lookup by ID succeeds, lookup missing ID returns false
- [X] T013 [P] [US1] Contract test: `POST /sessions` in `test/contract/session_api_test.go` — valid name → 201 unique ID; empty name → 400; name over 32 chars → 400; non-printable chars → 400; two same-name requests → distinct IDs
- [X] T014 [US1] Wire `POST /sessions` route in `cmd/server/main.go` using Go 1.22 `net/http` `ServeMux` with method+path pattern

---

## Phase 4: User Story 2 — Create and Join Games (P2)

**Story goal**: Sessions can create open games, join open games as second player, and cancel their own open games.

**Independent test criteria**: Create game → 201 `open`; join → 200 `active`; join own game → 403; join active game → 409; cancel by creator → 204; cancel by non-creator → 403.

- [X] T015 [P] [US2] Implement `GameStore` in `internal/store/game_store.go`: `sync.RWMutex`-protected map; methods `Create(g *domain.Game)`, `GetByID(id string) (*domain.Game, bool)`, `ListAll() []*domain.Game`, `Delete(id string)`, `JoinGame(gameID, joinerSessionID string) (*domain.Game, error)` — join uses write-lock compare-and-swap (check `Status == open`, assign joiner, initialize `chess.NewGame`, transition to `active`)
- [X] T016 [P] [US2] Implement session validation middleware in `internal/handler/middleware.go`: extract `X-Session-ID` header, look up session store, inject `*domain.Session` into request context; return 401 `missing_session_header` or 401 `session_not_found` on failure
- [X] T017 [US2] Implement `POST /games` handler in `internal/handler/game_handler.go`: require middleware, create `domain.Game` (UUID ID, `WhiteSessionID = session.ID`, `Status = open`), store, return `GameSummaryResponse` 201
- [X] T018 [US2] Implement `POST /games/{id}/join` handler in `internal/handler/join_handler.go`: require middleware, extract game ID from path, call `GameStore.JoinGame`; return 200 `GameDetailResponse` on success; return 404 `game_not_found`, 403 `cannot_join_own_game`, or 409 `game_not_open` on failure
- [X] T019 [US2] Implement `DELETE /games/{id}` handler in `internal/handler/game_handler.go`: require middleware, extract game ID, verify caller is `WhiteSessionID` and `Status == open`; call `GameStore.Delete`; return 204 on success; return 403 `cannot_cancel_game` or 404 `game_not_found` on failure
- [X] T020 [P] [US2] Unit test: `GameStore` in `test/unit/game_store_test.go` — create, list, get by ID, delete, join (success, double join returns error, join own game returns error, join non-open returns error)
- [X] T021 [P] [US2] Contract test: game lifecycle in `test/contract/game_lifecycle_test.go` — create game → open; join → active; join already-active → 409; join own → 403; cancel by creator → 204; cancel by non-creator → 403; missing session header → 401
- [X] T022 [US2] Wire `POST /games`, `DELETE /games/{id}`, `POST /games/{id}/join` routes in `cmd/server/main.go`

---

## Phase 5: User Story 3 — Play a Game (P3)

**Story goal**: Players can submit moves, resign, and claim draws in an active game; game status transitions to finished on terminal conditions.

**Independent test criteria**: Submit legal move → 200 accepted new state; submit illegal move → 422; wrong turn → 403; non-participant → 403; resign → 200 finished; checkmate sequence → finished correct winner; draw claim eligible → 200 finished draw; draw claim ineligible → 200 accepted=false.

- [X] T023 [P] [US3] Implement `POST /games/{id}/moves` handler in `internal/handler/move_handler.go`: require middleware; validate session is participant; determine active color (white=WhiteSessionID, black=BlackSessionID); verify caller's color matches `ChessGame.ActivePlayer()`; call `ChessGame.SubmitMove(notation)`; update `game.ChessGame` to new state; if terminal set `game.Status = finished`; return `MoveResponse`; error codes: 400 `invalid_notation`, 403 `not_a_participant`, 403 `wrong_turn`, 404 `game_not_found`, 409 `game_not_active`, 422 `illegal_move`
- [X] T024 [P] [US3] Implement `POST /games/{id}/resign` handler in `internal/handler/resign_handler.go`: require middleware; validate participant; determine resigning color; call `ChessGame.Resign(color)`; set `game.Status = finished`; return `ResignResponse`
- [X] T025 [P] [US3] Implement `POST /games/{id}/draw` handler in `internal/handler/draw_handler.go`: require middleware; validate participant and active player; validate `claim_type` string; call `ChessGame.ClaimDraw(claimType)`; if accepted set `game.Status = finished`; return `DrawResponse` with `accepted` bool and `rejection_reason` if rejected; error codes: 400 `invalid_draw_claim_type`, 403 `not_a_participant`, 403 `wrong_turn`, 409 `game_not_active`
- [X] T026 [P] [US3] Add `UpdateChessGame(gameID string, newGame *chess.Game, finished bool) error` method to `GameStore` in `internal/store/game_store.go` (used by move, resign, draw handlers to persist new chess state under write lock). **If a draw claim is rejected, `UpdateChessGame` is NOT called; the handler returns `DrawResponse{Accepted: false}` immediately without acquiring the write lock.)**
- [X] T027 [P] [US3] Contract test: move API in `test/contract/move_api_test.go` — legal move accepted; illegal move 422; wrong turn 403; non-participant 403; game not active 409; invalid notation 400; resign by participant 200 finished; resign by non-participant 403; draw claim ineligible 200 `accepted=false`; draw claim eligible (construct 50-move-rule position via FEN restore or sequential moves) 200 `accepted=true` + `Status=finished`; draw claim with unknown type 400
- [X] T028 [P] [US3] Integration test: full game flow in `test/integration/full_game_test.go` — create sessions → create game → join → play Fool's Mate (4 moves) → verify `finished` + `black_wins` outcome
- [X] T029 [US3] Wire `POST /games/{id}/moves`, `POST /games/{id}/resign`, `POST /games/{id}/draw` routes in `cmd/server/main.go`

---

## Phase 6: User Story 4 — List Available Games (P4)

**Story goal**: Any session can retrieve all games (all statuses) and get full detail for a specific game.

**Independent test criteria**: `GET /games` returns all; empty list when none; `GET /games/{id}` returns full FEN/history/outcome; 404 for missing ID.

- [X] T030 [P] [US4] Implement `GET /games` handler in `internal/handler/game_handler.go`: require middleware; call `GameStore.ListAll()`; map to `[]GameSummaryResponse`; return 200 with `{"games": [...]}`; return `{"games": []}` when empty
- [X] T031 [P] [US4] Implement `GET /games/{id}` handler in `internal/handler/game_handler.go`: require middleware; look up game; map to `GameDetailResponse` (include FEN from `ChessGame.ExportFEN()` if active/finished, null if open; include `active_color`, `move_history`, `outcome`); return 404 `game_not_found` if missing
- [X] T032 [P] [US4] Contract test: listing in `test/contract/game_lifecycle_test.go` (extend) — list returns all statuses; empty list; get by ID returns FEN and move history; get non-existent ID → 404; open game has null FEN; active game has valid FEN
- [X] T033 [US4] Wire `GET /games` and `GET /games/{id}` routes in `cmd/server/main.go`

---

## Phase 7: Concurrency Safety

**Story**: Ensures first-request-wins join under concurrent load (maps to SC-004, FR-015).

**Independent test criteria**: Two goroutines simultaneously join the same open game → exactly one 200 and one 409.

- [X] T034 Integration test: concurrent join in `test/integration/concurrent_join_test.go` — launch two goroutines simultaneously calling `POST /games/{id}/join`; assert exactly one 200 response and one 409 `game_not_open` response

---

## Phase 8: Polish & Cross-Cutting Concerns

- [X] T035 [P] Finalize `cmd/server/main.go`: instantiate stores, register all routes with method+path patterns, start `http.ListenAndServe` on `:8080` (port from env `PORT` with fallback)
- [X] T036 [P] Fix duplicate Clarifications section in `specs/001-chess-game-server/spec.md` (two identical `## Clarifications` blocks present)
- [X] T037 [P] Add `README.md` with build/run instructions: `go build ./cmd/server`, `./server`, environment variables, curl quickstart reference
- [X] T038 [P] Run `go mod tidy` and verify `go.mod` has no external dependencies beyond `ajedrez4-common`
- [X] T039 Run all unit tests: `go test ./test/unit/...`
- [X] T040 Run all contract tests: `go test ./test/contract/...`
- [X] T041 Run all integration tests: `go test ./test/integration/...`
- [X] T042 Run race detector: `go test -race ./...`
- [X] T043 Run `go vet ./...` and fix any issues
- [X] T044 Verify `go build ./...` produces no errors or warnings

---

## Dependencies

```
Phase 1 (T001–T004)
  └─► Phase 2 (T005–T009)  [all phases block on this]
        ├─► Phase 3 / US1 (T010–T014)  [independent]
        ├─► Phase 4 / US2 (T015–T022)  [requires US1 middleware T016]
        │     └─► Phase 5 / US3 (T023–T029)  [requires US2 game store T015, T026]
        │           └─► Phase 6 / US4 (T030–T033)  [requires US3 for FEN export in GET /games/{id}]
        └─► Phase 7 (T034)  [requires T015 + T018]
Phase 8 (T035–T044) runs after all phases complete
```

**Story independence**: US1 (sessions) can be built and tested before any other story. US2 requires the session middleware from US1's store. US3 requires the game store from US2. US4 requires the chess state from US3 (FEN export). US1 and US4 have no circular dependencies.

## Parallel Execution Examples

**Phase 2** — All of T005–T009 can be worked in parallel (different files, no intra-phase dependencies).

**Phase 3** — T010 (store) and T012–T013 (tests) can start in parallel; T011 (handler) requires T010.

**Phase 4** — T015 (store) and T016 (middleware) can start in parallel; T017–T019 (handlers) require both; T020–T021 (tests) require T015–T016.

**Phase 5** — T023–T026 can be worked in parallel (different handler files + store method); T027–T029 require them.

**Phase 6** — T030 and T031 are independent handler methods in the same file; T032 requires both.

**Phase 8** — T035–T038 are independent polish tasks; T039–T044 run sequentially after all code tasks.

## Implementation Strategy

**MVP (Phase 1 + 2 + 3)**: Just session creation. One endpoint, one store, one handler. Verifies the module setup, ajedrez4-common local replace, and HTTP wiring work end-to-end.

**Increment 2 (Phase 4)**: Add game creation + join + cancel. No chess logic yet — just lobby state management.

**Increment 3 (Phase 5)**: Add move submission, resign, draw. This is where ajedrez4-common integration is exercised fully.

**Increment 4 (Phase 6 + 7)**: Add listing endpoints and concurrent join safety test.

**Increment 5 (Phase 8)**: Polish, README, full test pass, race detector.
