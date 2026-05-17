## 1. Project Foundation

- [x] 1.1 Initialize Go module (`go mod init`) and create the directory tree: `cmd/server/`, `internal/config/`, `internal/model/`, `internal/platform/`, `internal/handler/`, `internal/middleware/`, `docs/`
- [x] 1.2 Write `internal/config/config_test.go`: assert all required env vars are loaded, that missing vars produce an error, and that no global state is mutated
- [x] 1.3 Implement `internal/config` package to pass tests: load all env vars (`BNET_CLIENT_ID`, `BNET_CLIENT_SECRET`, `STEAM_API_KEY`, `PSN_AUTH_TOKEN`, `XBOX_CLIENT_ID`, `XBOX_CLIENT_SECRET`, `JWT_SECRET`, `RATE_LIMIT_RPS`) with no global state
- [x] 1.4 Write `internal/model/model_test.go`: assert JSON serialization/deserialization round-trips for `PlayerStats`, `Character`, and `SeasonData` including edge cases (empty characters array, zero values)
- [x] 1.5 Define shared models in `internal/model` to pass tests: `PlayerStats`, `Character`, `SeasonData` structs with JSON tags

## 2. Platform Client Interface

- [x] 2.1 Write `internal/platform/client_test.go`: assert typed sentinel errors (`ErrPSNTokenInvalid`, `ErrPlatformRateLimited`, `ErrPlayerNotFound`) are distinct and implement the `error` interface
- [x] 2.2 Define `platform.Client` interface in `internal/platform/client.go` and typed sentinel errors to pass tests

## 3. Battle.net Client

- [x] 3.1 Write `internal/platform/battlenet/battlenet_test.go` with recorded HTTP fixtures covering: successful `FetchStats` (assert normalized `PlayerStats` field-by-field), token refresh on `401`, and `ErrPlatformRateLimited` on `429`
- [x] 3.2 Implement `internal/platform/battlenet` client: OAuth 2.0 client credentials token fetch with in-memory token cache (mutex-protected) to pass the token and rate-limit tests
- [x] 3.3 Implement `FetchStats` for Battle.net: call the Blizzard API endpoint and map the response to `*model.PlayerStats` to pass the normalization fixture test

## 4. Steam Client

- [x] 4.1 Write `internal/platform/steam/steam_test.go` with recorded HTTP fixtures covering: successful `FetchStats` (assert `platform: "steam"` and normalized fields), API key present on every request, and `ErrPlatformRateLimited` on `429`
- [x] 4.2 Implement `internal/platform/steam` client: attach `STEAM_API_KEY` as query param and map response to `*model.PlayerStats` to pass all Steam tests

## 5. PSN Client

- [x] 5.1 Write `internal/platform/psn/psn_test.go` with recorded HTTP fixtures covering: successful `FetchStats`, bearer token present on every request, and `ErrPSNTokenInvalid` returned on `401`
- [x] 5.2 Implement `internal/platform/psn` client: send `PSN_AUTH_TOKEN` as `Authorization: Bearer`, return `ErrPSNTokenInvalid` on `401`, and map response to `*model.PlayerStats` to pass all PSN tests

## 6. Xbox Client

- [x] 6.1 Write `internal/platform/xbox/xbox_test.go` with recorded HTTP fixtures covering: successful `FetchStats` (assert `platform: "xbox"` and normalized fields), XSTS token cached across calls, and token refresh on expiry
- [x] 6.2 Implement `internal/platform/xbox` client: OAuth 2.0 client credentials → XSTS token exchange with in-memory cache and map response to `*model.PlayerStats` to pass all Xbox tests

## 7. HTTP Handlers

- [x] 7.1 Write `internal/handler/handler_test.go` using a mock `platform.Client`: assert `GET /stats/{platform}/{username}` returns `200` with correct JSON body on success
- [x] 7.2 Extend handler tests: assert `400` for unknown platform, `404` for `ErrPlayerNotFound`, `502` for `ErrPSNTokenInvalid`, `502` for `ErrPlatformRateLimited`, and error envelope shape on all errors
- [x] 7.3 Extend handler tests: assert `GET /health` returns `200 {"status":"ok"}` without authentication
- [x] 7.4 Implement `GET /stats/{platform}/{username}` and `GET /health` handlers to pass all handler tests
- [x] 7.5 Register routes in `internal/handler/routes.go` and wire middleware chain (logging → rate limiting → auth, with health route bypassing auth)

## 8. Middleware

- [x] 8.1 Write `internal/middleware/auth_test.go`: assert valid JWT passes through, missing/expired/malformed tokens return `401` with correct error codes, and `GET /health` bypasses auth
- [x] 8.2 Implement JWT authentication middleware in `internal/middleware/auth.go` to pass all auth tests
- [x] 8.3 Write `internal/middleware/ratelimit_test.go`: assert requests within `RATE_LIMIT_RPS` pass, excess requests return `429` with error envelope, and limits are per source IP
- [x] 8.4 Implement per-IP rate limiting middleware in `internal/middleware/ratelimit.go` to pass all rate-limit tests
- [x] 8.5 Write `internal/middleware/logging_test.go`: assert every request produces a structured log entry containing `method`, `path`, `status`, and `latency_ms`, and that log output is valid JSON
- [x] 8.6 Implement request logging middleware in `internal/middleware/logging.go` using `log/slog` to pass all logging tests

## 9. Server Entrypoint

- [x] 9.1 Implement `cmd/server/main.go`: load config, construct platform clients, wire handler and middleware, start `net/http` server on port 8080
- [x] 9.2 Add graceful shutdown on `SIGINT`/`SIGTERM`

## 10. Documentation and Verification

- [x] 10.1 Write `docs/` per-platform auth setup instructions (how to obtain each credential)
- [x] 10.2 Run `go test ./...` and confirm all tests pass with no network calls
- [x] 10.3 Run `go vet ./...` and resolve any findings
- [x] 10.4 Smoke-test the running server: `go run ./cmd/server`, call `GET /health`, verify `200 {"status":"ok"}`
