## ADDED Requirements

### Requirement: Retrieve player stats by platform and username
The API SHALL expose a `GET /stats/{platform}/{username}` endpoint that returns normalized Diablo 4 player statistics for the given player. `{platform}` MUST be one of `battlenet`, `steam`, `psn`, or `xbox`. The response MUST conform to the unified `PlayerStats` schema regardless of which platform is queried.

#### Scenario: Successful stats retrieval
- **WHEN** a valid authenticated request is made to `GET /stats/battlenet/SomePlayer%231234`
- **THEN** the API returns `200 OK` with a JSON body matching the `PlayerStats` schema

#### Scenario: Unknown platform returns 400
- **WHEN** a request is made with a platform value not in the allowed set (e.g., `GET /stats/nintendo/foo`)
- **THEN** the API returns `400 Bad Request` with `{"error": "unsupported platform", "code": "INVALID_PLATFORM"}`

#### Scenario: Player not found on platform
- **WHEN** the platform API reports the requested username does not exist
- **THEN** the API returns `404 Not Found` with `{"error": "player not found", "code": "PLAYER_NOT_FOUND"}`

#### Scenario: Platform auth expired (PSN)
- **WHEN** the PSN token is expired or invalid
- **THEN** the API returns `502 Bad Gateway` with `{"error": "platform authentication expired", "code": "PSN_TOKEN_INVALID"}`

#### Scenario: Platform rate limited
- **WHEN** the upstream platform API returns a 429 response
- **THEN** the API returns `502 Bad Gateway` with `{"error": "platform rate limited", "code": "PLATFORM_RATE_LIMITED"}`

### Requirement: Unified PlayerStats response schema
The API SHALL return a consistent JSON schema for all platforms. The schema MUST include:
- `username` (string): the player's display name as returned by the platform
- `platform` (string): the platform identifier (`battlenet`, `steam`, `psn`, or `xbox`)
- `characters` (array of Character): list of the player's Diablo 4 characters
- `season` (SeasonData): current season progress data

Each `Character` MUST include: `name`, `class`, `level`, `paragon_level`.
`SeasonData` MUST include: `season_number`, `journey_progress`.

#### Scenario: Response includes all required fields
- **WHEN** a valid stats response is returned
- **THEN** the JSON body contains `username`, `platform`, `characters`, and `season` at the top level

#### Scenario: Characters array may be empty
- **WHEN** a player has no characters on the queried platform
- **THEN** `characters` is an empty array `[]`, not null or absent

### Requirement: Health check endpoint
The API SHALL expose `GET /health` that returns `200 OK` with `{"status": "ok"}`. This endpoint MUST NOT require authentication.

#### Scenario: Health check responds without auth
- **WHEN** an unauthenticated request is made to `GET /health`
- **THEN** the API returns `200 OK` with `{"status": "ok"}`

### Requirement: Consistent error envelope
All error responses from the API SHALL use the JSON envelope `{"error": "<human-readable message>", "code": "<machine-readable code>"}`.

#### Scenario: Error response has required fields
- **WHEN** any error condition is encountered
- **THEN** the response body contains both `error` and `code` string fields
