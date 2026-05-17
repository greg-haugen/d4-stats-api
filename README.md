# d4-stats-api

A Go-based REST API for retrieving Diablo 4 player statistics across multiple platforms. Given an authenticated player username, the API fetches and returns character stats, progression, and leaderboard data from Battle.net, Steam, PlayStation Network (PS5), and Xbox.

## Supported Platforms

| Platform | Auth Method |
|----------|-------------|
| Battle.net | OAuth 2.0 (Blizzard) |
| Steam | Steam Web API Key |
| PlayStation Network | PSN Auth Token |
| Xbox | Xbox Live / Microsoft OAuth |

## Features

- Unified response schema across all platforms
- Per-platform authentication handling
- Character stats, paragon levels, and season journey data
- Rate limiting and caching layer
- Structured logging and error responses

## Getting Started

### Prerequisites

- Go 1.22+
- API credentials for each platform you intend to query (see [docs/auth.md](docs/auth.md))

### Installation

```bash
git clone https://github.com/greg-haugen/d4-stats-api.git
cd d4-stats-api
go mod tidy
```

### Configuration

Copy the example env file and fill in your credentials:

```bash
cp .env.example .env
```

### Running the server

```bash
go run ./cmd/server
```

The API will start on `http://localhost:8080` by default.

## API Reference

### `GET /v1/player/{platform}/{username}`

Returns Diablo 4 stats for the given player.

**Path parameters:**

| Param | Values | Description |
|-------|--------|-------------|
| `platform` | `battlenet`, `steam`, `psn`, `xbox` | The platform to query |
| `username` | string | The player's username or identifier on that platform |

**Example:**

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/v1/player/battlenet/YourBattleTag%231234
```

**Response:**

```json
{
  "platform": "battlenet",
  "username": "YourBattleTag#1234",
  "characters": [...],
  "season": {...},
  "fetched_at": "2026-05-17T00:00:00Z"
}
```

## Project Structure

```
d4-stats-api/
├── cmd/
│   └── server/         # Main entrypoint
├── internal/
│   ├── platform/       # Per-platform clients (battlenet, steam, psn, xbox)
│   ├── handler/        # HTTP handlers
│   ├── middleware/      # Auth, rate limiting, logging
│   └── model/          # Shared response types
├── docs/               # Auth setup guides per platform
└── .env.example
```

## Contributing

Pull requests are welcome. Please open an issue first for any significant changes.

## License

MIT
