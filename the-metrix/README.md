# The Metrix

A comprehensive suite of developer productivity and team analytics tools.

## Overview

The Metrix is a modular platform for tracking developer metrics, team performance, and AI coding assistance. It consists of:

- **daily-metrics** - CLI tool for tracking daily work activities (commits, PRs, Jira issues)
- **metrics-api** - Go REST API for storing and querying metrics data
- **team-analytics-app** - Web dashboard for visualizing team sprint metrics and velocity
- **splash-screen-tester** - UI components for splash screens and feedback widgets

## Features

- Track individual developer productivity (commits, LOC, PRs)
- Monitor team sprint velocity and completion rates
- Visualize Jira issue progress and workload distribution
- Generate daily work summaries in markdown
- AI coding assistant metrics and time savings analytics
- RESTful API for integrating with other tools

## Quick Start

### 1. Environment Setup

Each component needs environment configuration. Copy the example files:

```bash
cd "The Metrix"

# Root configuration (optional - for shared settings)
cp .env.example .env

# Component-specific configs
cp daily-metrics/.env.example daily-metrics/.env
cp metrics-api/.env.example metrics-api/.env
cp team-analytics-app/.env.example team-analytics-app/.env
```

Edit each `.env` file with your organization's settings:

```bash
# Required: Jira Configuration
JIRA_BASE_URL=https://your-company.atlassian.net
JIRA_API_TOKEN=your_jira_api_token
JIRA_USER_EMAIL=your.email@company.com

# Required: GitHub Configuration  
GITHUB_TOKEN=your_github_token
GITHUB_API_URL=https://api.github.com  # or your GHE API URL

# Optional: Project/Org names
PROJECT_NAME=my-project
ORG_NAME=MyOrganization
TEAM_NAME=Engineering Team
```

### 2. Start Metrics API (Backend)

```bash
cd metrics-api

# Build the server
go build ./cmd/server

# Run (uses config.yaml + environment variables)
./server
```

API will be available at `http://localhost:8080`

### 3. Start Team Analytics App (Frontend)

```bash
cd team-analytics-app

# Install dependencies (using uv package manager)
uv sync

# Run development server
uv run uvicorn src.main:app --reload --port 8000
```

Dashboard will be available at `http://localhost:8000`

### 4. Use Daily Metrics CLI

```bash
cd daily-metrics

# Build the CLI
go build .

# Install globally (optional)
go install .

# Generate daily summary
daily-metrics --output ~/daily-notes/$(date +%Y-%m-%d).md

# Record a session note
daily-metrics record --note "Fixed ArgoCD auth" --jira PROJ-123 --pr 42
```

## Configuration

### Environment Variables

All components support these environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JIRA_BASE_URL` | Jira instance URL | `https://your-company.atlassian.net` | Yes |
| `JIRA_API_TOKEN` | Jira API token | - | Yes |
| `JIRA_USER_EMAIL` | Your Jira email | - | Yes |
| `GITHUB_TOKEN` | GitHub personal access token | - | Yes |
| `GITHUB_API_URL` | GitHub API endpoint | `https://api.github.com` | No |
| `METRICS_API_URL` | Metrics API base URL | `http://localhost:8080` | No |
| `PROJECT_NAME` | Default project name | `my-project` | No |
| `ORG_NAME` | Organization name | `MyOrganization` | No |
| `TEAM_NAME` | Team name | `Engineering Team` | No |

### Component-Specific Configuration

#### Daily Metrics CLI

Uses `~/.config/daily-metrics/config.json`:

```json
{
  "work_repos": [
    "/path/to/project1",
    "/path/to/project2"
  ],
  "private_repos": [
    "/path/to/personal-notes"
  ],
  "typing_speed_wpm": 50,
  "metrics_api_url": "http://localhost:8080"
}
```

#### Metrics API

Uses `config.yaml` in the metrics-api directory:

```yaml
repos:
  work:
    - main-project
    - platform-ops
  private:
    - personal-notes

cache:
  enabled: true
  ttl_seconds: 300

users:
  - e_number: "E12345"
    jira_account_id: "abc123"
    github_username: "youruser"
    name: "Your Name"
    email: "your.email@company.com"
```

#### Team Analytics App

Settings are configured via the web UI at `/settings`:

- **Jira Project Key** - Which project to track (e.g., PROJ, TEAM)
- **Workflow Type** - Agile (sprints) or Kanban (time-based)
- **Date Ranges** - Custom date ranges for analysis
- **Theme** - Dark or light mode
- **Demo Mode** - Toggle between live data and demo data

Sprint capacity planning can be manually configured in `src/config.py` if not using the API.

## Architecture

```
┌─────────────────────┐
│  daily-metrics CLI  │
│  (Go)               │
└──────────┬──────────┘
           │
           │ HTTP
           ▼
┌─────────────────────┐
│   metrics-api       │
│   (Go + SQLite)     │
└──────────┬──────────┘
           │
           │ HTTP
           ▼
┌─────────────────────┐
│ team-analytics-app  │
│ (Python/FastAPI)    │
└─────────────────────┘
```

- **daily-metrics** queries GitHub/Jira APIs, optionally submits to metrics-api
- **metrics-api** stores metrics in SQLite, provides REST endpoints
- **team-analytics-app** fetches from metrics-api or queries Jira directly (fallback)

## Development

### Prerequisites

- Go 1.21+ (for daily-metrics, metrics-api)
- Python 3.13+ (for team-analytics-app)
- [uv](https://github.com/astral-sh/uv) package manager (for Python)
- Jira API token ([create here](https://id.atlassian.com/manage-profile/security/api-tokens))
- GitHub personal access token ([create here](https://github.com/settings/tokens))

### Building from Source

#### Daily Metrics CLI

```bash
cd daily-metrics
go build .
./daily-metrics --help
```

#### Metrics API

```bash
cd metrics-api
go build ./cmd/server
./server
```

#### Team Analytics App

```bash
cd team-analytics-app
uv sync
uv run uvicorn src.main:app --reload
```

### Running Tests

```bash
# Go components
cd daily-metrics && go test ./...
cd metrics-api && go test ./...

# Python component (if tests exist)
cd team-analytics-app && uv run pytest
```

## API Documentation

### Metrics API Endpoints

- `GET /health` - Health check
- `GET /api/repos` - List tracked repositories
- `GET /api/metrics/daily` - Daily metrics summary
- `POST /api/opencode/submit` - Submit AI coding session metrics
- `GET /api/opencode/sessions` - Query AI coding sessions
- `GET /api/users` - List configured users

See `metrics-api/docs/` for full API documentation.

### Team Analytics App Endpoints

- `GET /` - Sprint metrics dashboard
- `GET /settings` - Configuration UI
- `GET /opencode` - AI metrics visualization
- `GET /api/status` - API/config mode status

## Deployment

### Docker (Recommended)

```bash
# Build metrics-api
cd metrics-api
docker build -t metrics-api:latest .
docker run -p 8080:8080 -v $(pwd)/metrics.db:/app/metrics.db metrics-api:latest

# Build team-analytics-app
cd team-analytics-app
docker build -t team-analytics:latest .
docker run -p 8000:8000 team-analytics:latest
```

### Systemd Service

Example service file for metrics-api:

```ini
[Unit]
Description=Metrics API
After=network.target

[Service]
Type=simple
User=metrics
WorkingDirectory=/opt/metrics-api
ExecStart=/opt/metrics-api/server
Restart=on-failure
Environment="JIRA_BASE_URL=https://company.atlassian.net"
Environment="JIRA_API_TOKEN=..."
EnvironmentFile=/etc/metrics-api/.env

[Install]
WantedBy=multi-user.target
```

## Troubleshooting

### "No Jira issues found"

- Check `JIRA_API_TOKEN` is valid
- Verify `JIRA_USER_EMAIL` matches your Atlassian account
- Ensure your Jira user has access to the project

### "GitHub rate limit exceeded"

- Use a GitHub personal access token (not anonymous)
- For GitHub Enterprise, set `GITHUB_API_URL` to your GHE instance

### "Metrics API connection failed"

- Check metrics-api is running: `curl http://localhost:8080/health`
- Verify `METRICS_API_URL` is set correctly
- Team analytics app will fall back to config mode if API is unavailable

### "Module not found" errors (Go)

- Run `go mod tidy` in the component directory
- Verify Go version is 1.21+

## License

[Your License Here]

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Support

For issues or questions:

- Open an issue on GitHub
- Check existing documentation in component READMEs
- Review example configurations in `.env.example` files
