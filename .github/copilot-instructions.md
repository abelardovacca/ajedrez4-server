<!-- SPECKIT START -->
## Implementation Plan: Chess Game Server

**Active Feature**: `001-chess-game-server` (v1.0.0)  
**Spec Status**: Complete (clarified)  
**Plan Status**: Phase 1 Design complete (research.md, data-model.md, contracts/, quickstart.md generated)

**Key Context**:
- Language: Go 1.22.2
- Architecture: Layered (domain → store → handler → api)
- Protocol: HTTP REST, JSON payloads
- Session auth: `X-Session-ID` request header
- Storage: In-memory (`sync.RWMutex`-protected maps), no persistence
- Chess logic: delegated entirely to `github.com/ajedrez4/ajedrez4-common` (local replace directive)
- Concurrency: First-request-wins join via write-lock compare-and-swap
- Dependencies: Zero external deps beyond ajedrez4-common
- Color assignment: creator = white, joiner = black

**Design Artifacts**:
- [specs/001-chess-game-server/plan.md](specs/001-chess-game-server/plan.md) - Implementation plan with technical context and project structure
- [specs/001-chess-game-server/research.md](specs/001-chess-game-server/research.md) - Phase 0 research (all decisions resolved)
- [specs/001-chess-game-server/data-model.md](specs/001-chess-game-server/data-model.md) - Entity definitions and state transitions
- [specs/001-chess-game-server/contracts/http-api.md](specs/001-chess-game-server/contracts/http-api.md) - Full HTTP REST API contract
- [specs/001-chess-game-server/quickstart.md](specs/001-chess-game-server/quickstart.md) - Usage examples (curl)

**Next Step**: Run `/speckit.tasks` to generate implementation tasks
<!-- SPECKIT END -->
