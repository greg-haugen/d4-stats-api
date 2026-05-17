## ADDED Requirements

### Requirement: JWT authentication middleware
The API SHALL validate a JWT bearer token on all routes except `GET /health`. The JWT MUST be signed with a secret loaded from the `JWT_SECRET` environment variable using HMAC-SHA256. Requests with a missing, malformed, or expired token MUST be rejected before reaching any handler.

#### Scenario: Valid JWT allows request through
- **WHEN** a request includes a valid, non-expired JWT in the `Authorization: Bearer <token>` header
- **THEN** the middleware passes the request to the next handler

#### Scenario: Missing JWT returns 401
- **WHEN** a request to a protected route has no `Authorization` header
- **THEN** the middleware returns `401 Unauthorized` with `{"error": "missing token", "code": "MISSING_TOKEN"}`

#### Scenario: Expired JWT returns 401
- **WHEN** a request includes a JWT whose `exp` claim is in the past
- **THEN** the middleware returns `401 Unauthorized` with `{"error": "token expired", "code": "TOKEN_EXPIRED"}`

#### Scenario: Malformed JWT returns 401
- **WHEN** a request includes a syntactically invalid token string
- **THEN** the middleware returns `401 Unauthorized` with `{"error": "invalid token", "code": "INVALID_TOKEN"}`

#### Scenario: Health check bypasses auth
- **WHEN** an unauthenticated request is made to `GET /health`
- **THEN** the JWT middleware does not reject the request

### Requirement: Rate limiting middleware
The API SHALL enforce per-IP rate limiting. The limit MUST be configurable via the `RATE_LIMIT_RPS` environment variable (requests per second). Requests exceeding the limit MUST be rejected with `429 Too Many Requests`.

#### Scenario: Requests within limit are allowed
- **WHEN** a client sends requests at a rate at or below `RATE_LIMIT_RPS`
- **THEN** all requests reach the handler

#### Scenario: Requests exceeding limit are rejected
- **WHEN** a client exceeds `RATE_LIMIT_RPS` requests per second
- **THEN** the middleware returns `429 Too Many Requests` with `{"error": "rate limit exceeded", "code": "RATE_LIMITED"}`

#### Scenario: Rate limit is per source IP
- **WHEN** two different client IPs each send requests at `RATE_LIMIT_RPS`
- **THEN** neither client is rate-limited (limits are not shared across IPs)

### Requirement: Request logging middleware
The API SHALL log every request and its outcome using `log/slog` with structured JSON output. Each log entry MUST include: HTTP method, path, response status code, and request latency in milliseconds.

#### Scenario: Every request produces a log line
- **WHEN** any request is handled (success or error)
- **THEN** a structured JSON log entry is written containing `method`, `path`, `status`, and `latency_ms`

#### Scenario: Log output is valid JSON
- **WHEN** the server is running with default config
- **THEN** each log line emitted to stdout is a valid JSON object
