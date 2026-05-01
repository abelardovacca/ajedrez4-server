# Research: Chess Game Server

**Feature**: `001-chess-game-server`  
**Phase**: 0 — Technology & Design Research  
**Date**: 2026-05-01

## Decision 1: Language and Runtime

**Decision**: Go 1.22.2  
**Rationale**: Same language as `ajedrez4-common`, confirmed available on the target system. Consistent toolchain, no context switching, same test runner (`go test`).  
**Alternatives considered**: Node.js, Python — rejected to avoid mixing language stacks across sibling repositories.

## Decision 2: HTTP Routing

**Decision**: `net/http` standard library (Go 1.22 enhanced ServeMux)  
**Rationale**: Go 1.22 introduced method-specific routing and path parameter extraction (`{id}`) directly into `net/http`. This eliminates any external router dependency. The API surface (≤9 endpoints) is well within the comfort zone of the stdlib mux.  
**Alternatives considered**:
- `go-chi/chi` — popular lightweight router; rejected because Go 1.22 stdlib covers the same patterns.
- `gin-gonic/gin` — full framework; rejected as over-engineered for this scope.
- `gorilla/mux` — archived; rejected.

## Decision 3: JSON Serialization

**Decision**: `encoding/json` standard library  
**Rationale**: No schema evolution or performance constraints specified for v1. Stdlib JSON is sufficient.  
**Alternatives considered**: `json-iterator`, `sonic` — rejected; no measurable performance target defined.

## Decision 4: ajedrez4-common Integration

**Decision**: Local module replacement via `go.mod replace` directive  
**Rationale**: `ajedrez4-common` lives in a sibling directory. Using a `replace` directive (`replace github.com/ajedrez4/ajedrez4-common => ../ajedrez4-common`) enables local development without publishing. The directive is removed when a versioned release is tagged.  
**Pattern**:
```
require github.com/ajedrez4/ajedrez4-common v0.0.0
replace github.com/ajedrez4/ajedrez4-common => ../ajedrez4-common
```

## Decision 5: Concurrency Model

**Decision**: `sync.RWMutex`-protected in-memory maps  
**Rationale**: In-memory v1 storage with concurrent HTTP requests requires explicit locking. `sync.RWMutex` allows multiple concurrent reads (game list, game detail) while serializing writes (create, join, move). Critical section for join uses a compare-and-swap pattern: check game status, then set under write lock, ensuring first-writer-wins semantics.  
**Alternatives considered**: Channel-based actor model — rejected as over-complex for the scale and scope.

## Decision 6: Session Token Transport

**Decision**: `X-Session-ID` request header  
**Rationale**: Clean REST convention. Avoids URL pollution. Simple to consume from any HTTP client. Session IDs are validated by looking up the store on each request (no cryptographic tokens needed for v1 — no auth).  
**Alternatives considered**: Bearer token in Authorization header — rejected; implies auth semantics not present in v1. Cookie — rejected; not idiomatic for programmatic API.

## Decision 7: Testing

**Decision**: `net/http/httptest` for handler tests, table-driven subtests via `testing` stdlib  
**Rationale**: `httptest.NewRecorder` and `httptest.NewServer` allow full HTTP handler testing without binding to a real port. No external test framework required.  
**Alternatives considered**: `testify` — not needed; stdlib `testing` with `t.Fatal` / `t.Errorf` is sufficient.

## Decision 8: Error Response Format

**Decision**: Structured JSON `{"error": {"code": "...", "message": "..."}}` with appropriate HTTP status codes  
**Rationale**: Machine-readable `code` field enables client switch/case logic. Human-readable `message` field aids debugging. Consistent envelope across all error responses.

| HTTP Status | Usage |
|-------------|-------|
| 400 | Validation errors (bad display name, bad notation) |
| 401 | Missing or invalid session token |
| 403 | Forbidden action (wrong player, not participant) |
| 404 | Resource not found (game ID, session ID) |
| 409 | Conflict (game already joined, already finished) |
| 422 | Illegal move (valid notation, illegal position) |

## Decision 9: Module Name

**Decision**: `github.com/ajedrez4/ajedrez4-server`  
**Rationale**: Consistent with the naming convention of the sibling repository `github.com/ajedrez4/ajedrez4-common`.
