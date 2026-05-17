## Context

`d4-stats-api` is a greenfield Go service. There is no existing code to migrate. The four target platforms (Battle.net, Steam, PSN, Xbox Live) each require different auth flows and return different JSON shapes; the central challenge is abstracting those differences behind a single interface without leaking platform details into the HTTP layer.

## Goals / Non-Goals

**Goals:**
- Deliver a working REST API that returns normalized D4 player stats for any of the four supported platforms
- Keep the HTTP handlers completely platform-agnostic via the `platform.Client` interface
- Ship with JWT auth and rate limiting from day one — security is not an afterthought
- Ensure CI never hits real platform APIs (HTTP fixture tests per platform)

**Non-Goals:**
- Persistent storage, user accounts, or session management
- Leaderboard writes or any mutating platform calls
- Support for platforms beyond Battle.net, Steam, PSN, and Xbox Live
- Redis caching (in-memory caching may be added later but is not required now)

## Decisions

### 1. Single `platform.Client` interface

All four platform packages expose the same `FetchStats(ctx context.Context, username string) (*model.PlayerStats, error)` method. Handlers receive the interface, not a concrete type. This keeps routing logic simple: look up the platform from the request, dispatch to the right client, serialize the result.

**Alternatives considered:** A union-type approach where handlers switch on platform and call different methods — rejected because it couples the HTTP layer to platform details and complicates testing.

### 2. OAuth client-credentials flow for Battle.net and Xbox

Both platforms use OAuth 2.0 client credentials. The token is fetched at startup and refreshed on 401. A simple in-process token cache (mutex-protected struct) is sufficient — no external cache needed for a stateless service.

**Alternatives considered:** Fetching a new token per request — rejected due to unnecessary latency and rate-limit exposure on the token endpoint.

### 3. PSN npsso token as a pre-issued credential

PSN does not expose a server-side OAuth flow suitable for client credentials. The operator supplies a `PSN_AUTH_TOKEN` env var. The PSN client treats this as a bearer token and refreshes it only on explicit failure. Operators must rotate this credential manually (documented in `docs/`).

**Alternatives considered:** Full PSN OAuth redirect flow — out of scope for a server-to-server API with no user sessions.

### 4. JWT for API caller authentication (not platform auth)

Callers of this API authenticate with a JWT signed by a secret the operator controls. Middleware validates the token before dispatching to any platform client. This decouples caller identity from platform credentials entirely.

**Alternatives considered:** API key header — JWT was chosen to allow standard claims (exp, sub) without a token store.

### 5. Environment-variable config, no global state

All credentials and tunables are read via a single `internal/config` package at startup. No `init()` functions or package-level vars hold config. This makes unit tests straightforward and avoids subtle ordering bugs.

### 6. `log/slog` with structured JSON output

Structured logs make it easy to ship to any log aggregator without a custom format. The request-logging middleware attaches method, path, status, and latency to every log line.

## Risks / Trade-offs

- **PSN token expiry** → Short-lived npsso tokens will cause PSN calls to fail until the operator rotates `PSN_AUTH_TOKEN`. Mitigate by returning a clear `{"error": "platform_auth_expired", "code": "PSN_TOKEN_INVALID"}` so callers can surface the issue.
- **Platform API shape changes** → If Battle.net, Steam, PSN, or Xbox change their response schema, the corresponding client's mapper breaks silently. Mitigate with strict fixture tests that pin the expected normalized output against a known raw response.
- **Rate limits from upstream platforms** → Heavy query traffic may exhaust platform API quotas. Mitigate by surfacing platform 429s as `{"error": "platform_rate_limited"}` and adding in-memory per-username caching later if needed.
- **JWT secret rotation** → No key rotation is built in. Mitigate by documenting the restart procedure and flagging this as a future enhancement.

## Migration Plan

This is a new service with no prior deployment. Steps:
1. Set all required environment variables (see Platform Auth Summary in CLAUDE.md).
2. Run `go run ./cmd/server` or build a binary and run it.
3. Verify health endpoint responds before routing traffic.

Rollback: stop the process; no persistent state to unwind.

## Open Questions

- Should the stats endpoint accept a `platform` query param or a URL segment (e.g., `/stats/{platform}/{username}`)? URL segment is more RESTful and is the assumed design.
- Is a `/health` endpoint required by any upstream load balancer? Assumed yes — include a simple `GET /health` returning `200 OK`.
