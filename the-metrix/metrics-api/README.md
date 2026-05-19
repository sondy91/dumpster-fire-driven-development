# Metrics API

Go REST API for team productivity metrics and analytics.

## Overview

Centralized API for querying:

- Jira issues and sprint data
- GitHub PRs and commits
- Git repository statistics (commits, LOC)
- Working days calculations (holidays, Schedule B)

Powers both the daily-metrics CLI and team-analytics web app.

## Architecture

```
metrics-api/
├── cmd/server/          # API server entrypoint
├── internal/
│   ├── api/             # HTTP handlers and routes
│   ├── clients/         # Jira/GitHub/Git client wrappers
│   ├── models/          # Data structures
│   ├── storage/         # SQLite database layer
│   └── config/          # Configuration management
├── pkg/                 # Shared libraries (working days, Git analysis)
├── docs/                # API documentation
└── config.yaml          # Service configuration
```

## Quick Start

```bash
# Build
go build ./cmd/server

# Run (starts on port 8080)
./server

# Run with custom config
CONFIG_PATH=/path/to/config.yaml ./server

# Test
go test ./...
```

Server will start on <http://localhost:8080> by default.

## API Endpoints

See [docs/api.md](docs/api.md) for full API documentation.

**Core Endpoints:**

- `GET /api/v1/health` - Health check
- `GET /api/v1/personal/daily` - Individual daily metrics
- `GET /api/v1/personal/summary` - Individual period summary
- `GET /api/v1/git/commits` - Git commit analysis
- `GET /api/v1/github/prs` - GitHub PR metrics
- `GET /api/v1/jira/issues` - Jira issue metrics

## Configuration

Configuration via `config.yaml` or environment variables:

```yaml
repos:
  work:
    - platform-ops
    - main-project
  private:
    - personal-notes

schedule_b_anchor: "2026-04-03"
holidays_2025: [...]
holidays_2026: [...]

api:
  port: 8080
  host: "0.0.0.0"

cache:
  enabled: true
  ttl_seconds: 300  # 5 minutes
```

## Development

**Prerequisites:**

- Go 1.21+
- Access to Jira CLI (`jira`)
- Access to GitHub CLI (`gh`)
- Git configured with user credentials

**Running locally:**

```bash
go run ./cmd/server
```

**Building:**

```bash
go build -o metrics-api ./cmd/server
```

**Testing:**

```bash
go test -v ./...
```

## Features

### Caching Layer

Metrics API implements a two-tier caching strategy:

1. **In-memory cache** - Fast access for repeated queries
2. **SQLite persistent cache** - Survives restarts, configurable TTL

Cache keys include query parameters (dates, e-numbers) for proper isolation. Configurable via `config.yaml`:

```yaml
cache:
  enabled: true
  ttl_seconds: 300  # Cache lifetime in seconds
```

### Smart Fallback

Client libraries (like daily-metrics CLI) automatically fall back to direct API calls if metrics-api is unavailable.

## Deployment

TBD - Will deploy to Kubernetes alongside team-analytics-app.

**Current Status:** Local development mode (localhost:8080)

## Related Projects

- **daily-metrics CLI** (`../daily-metrics/`) - Command-line client
- **team-analytics-app** (`../../src/team-analytics-app/`) - Web dashboard
