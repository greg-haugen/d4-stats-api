## Why

Diablo 4 player stats are spread across four separate platform APIs (Battle.net, Steam, PSN, Xbox Live), each with its own auth scheme and data shape. This API unifies them behind a single REST interface so callers never need to know which platform a player uses.

## What Changes

- New Go REST API server (`cmd/server`) serving player stats over HTTP on port 8080
- Four platform client implementations behind a shared `platform.Client` interface, each handling its own auth and data normalization
- Unified `PlayerStats` response schema (models: `PlayerStats`, `Character`, `SeasonData`) returned regardless of platform
- JWT authentication middleware protecting all routes
- Per-IP rate limiting middleware
- Structured JSON request logging via `log/slog`
- Environment-based config loading with no global state

## Capabilities

### New Capabilities

- `player-stats`: REST endpoints for retrieving a player's Diablo 4 stats, including the unified response schema and per-platform query routing
- `platform-integration`: The `platform.Client` interface and four implementations (Battle.net OAuth, Steam API key, PSN npsso, Xbox OAuth/XSTS), each normalizing raw platform data into the shared model
- `api-middleware`: JWT authentication, rate limiting, and structured request logging applied globally to all routes

### Modified Capabilities

## Impact

- Introduces the entire project structure under `cmd/`, `internal/`, and `docs/`
- Requires four sets of platform credentials as environment variables (`BNET_CLIENT_ID`, `BNET_CLIENT_SECRET`, `STEAM_API_KEY`, `PSN_AUTH_TOKEN`, `XBOX_CLIENT_ID`, `XBOX_CLIENT_SECRET`) plus a JWT signing secret
- No database or external storage; stateless service with optional in-memory caching path later
- Each platform package needs HTTP fixture tests — no live API calls in CI
