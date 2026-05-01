# Implementation Plan: Chess Game Server

**Branch**: `001-chess-game-server` | **Date**: 2026-05-01 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/001-chess-game-server/spec.md`

## Summary

Build an HTTP REST server in Go 1.22 that allows users to create sessions, create and join chess games, and play against each other. All chess rule enforcement is delegated to `ajedrez4-common`. Game state is stored in-memory with `sync.RWMutex`-protected maps. The server exposes 9 JSON endpoints. No authentication — any caller with network access can create a session.

## Technical Context

**Language/Version**: Go 1.22.2  
**Primary Dependencies**: `github.com/ajedrez4/ajedrez4-common` (local replace), stdlib only (`net/http`, `encoding/json`, `sync`, `crypto/rand`)  
**Storage**: In-memory (`sync.RWMutex`-protected maps); no persistence across restarts  
**Testing**: `go test`, `net/http/httptest`, table-driven subtests  
**Target Platform**: Linux server (amd64)  
**Project Type**: HTTP REST web service  
**Performance Goals**: None specified for v1 (in-memory, low concurrency expected)  
**Constraints**: Zero external dependencies beyond `ajedrez4-common`; no auth; polling model only  
**Scale/Scope**: Small number of concurrent games; single process, single binary

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution is a blank template (no project-specific principles have been ratified for ajedrez4-server). No gates to enforce. Architecture decisions are guided by the feature spec and research.md.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── server/
    └── main.go                  # Entry point: wire router, stores, start HTTP server

internal/
├── domain/
│   ├── session.go               # Session entity
│   ├── game.go                  # Game lobby entity + status transitions
│   └── errors.go                # Domain error codes and types
├── store/
│   ├── session_store.go         # In-memory session store (sync.RWMutex)
│   └── game_store.go            # In-memory game store (sync.RWMutex)
├── handler/
│   ├── session_handler.go       # POST /sessions
│   ├── game_handler.go          # GET /games, POST /games, GET /games/{id}, DELETE /games/{id}
│   ├── join_handler.go          # POST /games/{id}/join
│   ├── move_handler.go          # POST /games/{id}/moves
│   ├── resign_handler.go        # POST /games/{id}/resign
│   ├── draw_handler.go          # POST /games/{id}/draw
│   └── middleware.go            # Session validation middleware
└── api/
    ├── requests.go              # JSON request DTOs
    └── responses.go             # JSON response DTOs + error envelope

test/
├── unit/
│   ├── session_store_test.go
│   ├── game_store_test.go
│   └── domain_test.go
├── contract/
│   ├── session_api_test.go
│   ├── game_lifecycle_test.go
│   └── move_api_test.go
└── integration/
    ├── full_game_test.go
    └── concurrent_join_test.go

go.mod
go.sum
README.md
```

**Structure Decision**: Single Go module, clean layered architecture. `internal/domain` holds pure types, `internal/store` holds concurrency-safe state, `internal/handler` holds HTTP concerns, `internal/api` holds DTOs. `test/` mirrors ajedrez4-common conventions.

## Complexity Tracking

No constitution violations. No complexity justification required.
