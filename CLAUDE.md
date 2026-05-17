# CLAUDE.md

## Project Overview

`d4-stats-api` is a Go REST API that retrieves Diablo 4 player stats from four platforms: Battle.net, Steam, PlayStation Network (PSN), and Xbox Live. Each platform has its own authentication mechanism and data shape; the API normalizes these into a unified response schema.

## Architecture

```
cmd/server/          → main.go, server startup, env loading
internal/
  platform/          → one package per platform (battlenet, steam, psn, xbox)
                       each exposes a Client with a FetchStats(username) method
  handler/           → HTTP handlers, route registration
  middleware/        → JWT auth, rate limiting, request logging
  model/             → shared types: PlayerStats, Character, SeasonData
docs/                → per-platform auth setup instructions
```

The platform clients are behind a `platform.Client` interface so handlers stay platform-agnostic.

## Development Commands

```bash
go run ./cmd/server          # start dev server (port 8080)
go test ./...                # run all tests
go vet ./...                 # static analysis
golangci-lint run            # lint (requires golangci-lint installed)
```

## Key Conventions

- **Error responses** use a consistent `{"error": "...", "code": "..."}` JSON envelope.
- **Logging** uses `log/slog` with structured JSON output.
- **Config** is loaded from environment variables via a single `internal/config` package — no global state.
- **No ORM** — this service is stateless (no database); caching is in-memory or Redis if added later.
- Each platform package should have its own `_test.go` with a recorded HTTP fixture so tests never hit real APIs.

## Platform Auth Summary

| Platform | Credential needed | Notes |
|----------|------------------|-------|
| Battle.net | `BNET_CLIENT_ID`, `BNET_CLIENT_SECRET` | Blizzard OAuth client credentials flow |
| Steam | `STEAM_API_KEY` | Steam Web API key |
| PSN | `PSN_AUTH_TOKEN` | PSN npsso token, short-lived — may need refresh logic |
| Xbox | `XBOX_CLIENT_ID`, `XBOX_CLIENT_SECRET` | Microsoft OAuth, returns XSTS token |

## Out of Scope (for now)

- Persistent storage / user accounts
- Write operations (no leaderboard submission)
- Platforms beyond the four listed above
