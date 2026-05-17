## ADDED Requirements

### Requirement: Platform Client interface
The system SHALL define a `platform.Client` interface with a single method: `FetchStats(ctx context.Context, username string) (*model.PlayerStats, error)`. All platform packages MUST implement this interface. Handlers MUST depend only on this interface, never on concrete platform types.

#### Scenario: Handler uses interface not concrete type
- **WHEN** a handler is given a `platform.Client` for any of the four platforms
- **THEN** the handler can call `FetchStats` without knowing which concrete implementation it received

### Requirement: Battle.net client
The Battle.net client SHALL authenticate using the OAuth 2.0 client credentials flow with `BNET_CLIENT_ID` and `BNET_CLIENT_SECRET`. The client MUST cache the access token in memory and refresh it on `401 Unauthorized` responses from the Battle.net API.

#### Scenario: Token is reused within its validity window
- **WHEN** two `FetchStats` calls are made within the token's TTL
- **THEN** only one token fetch is performed against the Battle.net OAuth endpoint

#### Scenario: Token is refreshed on 401
- **WHEN** the Battle.net API returns `401 Unauthorized`
- **THEN** the client fetches a new token and retries the original request once

#### Scenario: Battle.net data is normalized
- **WHEN** `FetchStats` returns successfully for a Battle.net player
- **THEN** the returned `*model.PlayerStats` contains `platform: "battlenet"` and fields populated from Battle.net response data

### Requirement: Steam client
The Steam client SHALL authenticate using the `STEAM_API_KEY` environment variable as a query parameter on Steam Web API requests. No token refresh is required.

#### Scenario: Steam API key is sent on every request
- **WHEN** a Steam `FetchStats` call is made
- **THEN** the HTTP request to the Steam Web API includes the `key` query parameter

#### Scenario: Steam data is normalized
- **WHEN** `FetchStats` returns successfully for a Steam player
- **THEN** the returned `*model.PlayerStats` contains `platform: "steam"`

### Requirement: PSN client
The PSN client SHALL use the `PSN_AUTH_TOKEN` environment variable as a bearer token. The client MUST return a typed error distinguishing PSN token expiry from other failures.

#### Scenario: PSN token sent as bearer
- **WHEN** a PSN `FetchStats` call is made
- **THEN** the HTTP request includes `Authorization: Bearer <PSN_AUTH_TOKEN>`

#### Scenario: Expired PSN token produces typed error
- **WHEN** the PSN API returns `401 Unauthorized`
- **THEN** `FetchStats` returns an error that the handler can identify as `PSN_TOKEN_INVALID`

### Requirement: Xbox client
The Xbox client SHALL authenticate using the OAuth 2.0 client credentials flow with `XBOX_CLIENT_ID` and `XBOX_CLIENT_SECRET`, exchanging the resulting token for an XSTS token. The client MUST cache the XSTS token and refresh on expiry.

#### Scenario: XSTS token is cached
- **WHEN** two `FetchStats` calls are made within the XSTS token's TTL
- **THEN** only one XSTS token exchange is performed

#### Scenario: Xbox data is normalized
- **WHEN** `FetchStats` returns successfully for an Xbox player
- **THEN** the returned `*model.PlayerStats` contains `platform: "xbox"`

### Requirement: Fixture-based platform tests
Each platform client package SHALL include a `_test.go` file with recorded HTTP fixtures. Tests MUST NOT make real network calls to any external API. The fixture MUST pin the expected `*model.PlayerStats` output against a known raw platform response.

#### Scenario: Platform test runs offline
- **WHEN** `go test ./internal/platform/...` is run with no network access
- **THEN** all tests pass without connecting to any external host

#### Scenario: Fixture output is exact
- **WHEN** the recorded fixture response is fed to `FetchStats`
- **THEN** the returned `*model.PlayerStats` matches the expected normalized value field-by-field
