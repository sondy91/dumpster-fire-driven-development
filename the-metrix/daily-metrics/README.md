# daily-metrics

CLI tool for tracking personal productivity metrics including Jira issues, GitHub PRs, Git commits, and AI time savings.

## Installation

```bash
cd /path/to/The\ Metrix/daily-metrics
go build
go install
```

## Usage

### Default Command - Daily Report

Generate a daily productivity report:

```bash
daily-metrics [DATE]
```

Example:

```bash
daily-metrics 2026-04-17
daily-metrics  # defaults to today
```

**Options:**

- `--all` - Include all repos (public and private)
- `--output <file>` - Output markdown to file (e.g., daily note path)
- `--api-url <url>` - Metrics API base URL (default: <http://localhost:8080>)
- `--force-offline` - Skip API and use direct queries
- `--verbose` - Show API request details

**API Mode:**

The tool automatically tries to use the metrics-api if available, falling back to direct Jira/GitHub/Git queries if the API is down.

```bash
# Use API at default URL
daily-metrics

# Use API at custom URL
daily-metrics --api-url http://metrics-api:8080

# Force offline mode (skip API)
daily-metrics --force-offline

# Verbose API requests
daily-metrics --verbose
```

### `record` - Log Work Items

Record work items for later aggregation:

```bash
daily-metrics record --note "Description" [--jira XAIP-123] [--pr 42] [--no-timestamp]
```

**Options:**

- `--note` - Description of work done
- `--jira` - Jira issue ID (e.g., XAIP-123)
- `--pr` - Pull request number
- `--no-timestamp` - Omit timestamp from note

### `summary` - Multi-day Summary

Generate a summary report across multiple days, including OpenCode AI assistance metrics:

```bash
daily-metrics summary [--start YYYY-MM-DD] [--end YYYY-MM-DD] [--all] [--output file.md]
```

**Options:**

- `--start` - Start date (YYYY-MM-DD)
- `--end` - End date (YYYY-MM-DD)
- `--all` - Include all repos (work + private)
- `--merged-only` - Count only commits merged to main/default branch
- `--output` - Output markdown to file

The summary automatically includes:

- GitHub commits and LOC
- Pull requests (created and merged)
- Jira issues (worked on and completed)
- **OpenCode AI assistance** (commands, edits, time saved)

### `opencode` - AI Time Savings

Track productivity gains from OpenCode AI-assisted command execution.

```bash
daily-metrics opencode [--days N] [--start YYYY-MM-DD] [--end YYYY-MM-DD] [--wpm N] [--format FORMAT] [--output FILE]
```

**Options:**

- `--days` - Number of days to analyze (default: 7)
- `--start` - Start date (YYYY-MM-DD)
- `--end` - End date (YYYY-MM-DD)
- `--wpm` - Override typing speed (words per minute)
- `--format` - Output format: `terminal`, `json`, `markdown` (default: terminal)
- `--output` - Output file path (required for json/markdown formats)

**Examples:**

```bash
# Last 7 days (default, terminal output)
daily-metrics opencode

# Last 30 days
daily-metrics opencode --days 30

# Export to JSON for API integration
daily-metrics opencode --days 30 --format json --output stats.json

# Markdown report
daily-metrics opencode --days 30 --format markdown --output report.md

# Specific date range
daily-metrics opencode --start 2026-04-01 --end 2026-04-17

# Custom typing speed
daily-metrics opencode --wpm 40
```

**How it works:**

Analyzes bash commands and file edits executed by OpenCode and calculates time saved based on your typing speed.

Formula: `time_saved = total_chars / (wpm * 5 / 60)`

Where:

- `wpm` = words per minute (default: 50)
- `5` = average characters per word
- `60` = seconds per minute

The tool queries OpenCode's SQLite database at `~/.local/share/opencode/opencode.db` to extract:

- All completed bash commands
- All completed file edits (code generation)

It then counts the characters and estimates how long it would have taken you to type them manually.

**Features:**

- **Command typing saved** - Time saved from AI-generated bash commands
- **Code generation saved** - Time saved from AI-assisted file edits
- **Project breakdown** - See which repositories benefit most from AI assistance
- **Session tracking** - Understand AI usage patterns across work sessions

## Integration Tools

### `transform-to-api` - Export to Metrics API

Transform OpenCode aggregate statistics into metrics-api format for team dashboard integration.

**Location:** `cmd/transform-to-api/`

**Usage:**

```bash
cd tools/daily-metrics
go build -o transform-to-api ./cmd/transform-to-api

# Transform and submit to API
./transform-to-api --input stats.json --sprint "Sprint 12" --developer e40057167

# Dry run (preview without submitting)
./transform-to-api --input stats.json --sprint "Sprint 12" --dry-run

# Save transformed JSON to file
./transform-to-api --input stats.json --sprint "Sprint 12" --output transformed.json

# Custom API URL
./transform-to-api --input stats.json --sprint "Sprint 12" --api-url http://api.example.com:8080
```

**Options:**

- `--input` - Input JSON file from `daily-metrics opencode --format json` (required)
- `--sprint` - Sprint identifier (e.g., "Sprint 12") (required)
- `--developer` - Developer identifier (default: $USER)
- `--api-url` - Metrics API base URL (default: <http://localhost:8080>)
- `--output` - Save transformed JSON to file instead of submitting to API
- `--dry-run` - Preview what would be submitted without actually submitting

**Complete Workflow:**

```bash
# 1. Generate OpenCode stats as JSON
daily-metrics opencode --days 30 --format json --output opencode-stats.json

# 2. Transform and submit to metrics-api
cd tools/daily-metrics
go build -o transform-to-api ./cmd/transform-to-api
./transform-to-api --input opencode-stats.json --sprint "Sprint 12"
```

**What it does:**

- Reads aggregate OpenCode statistics (session count, tool breakdown, project summaries)
- Splits data by project (creates one API record per project)
- Distributes tool usage (bash, reads, edits, searches) proportionally across projects
- Calculates time savings per tool type
- Submits to metrics-api `/api/v1/opencode/submit` endpoint

**Data Format:**

The tool transforms from daily-metrics aggregate format:

```json
{
  "session_count": 75,
  "total_commands": 13954,
  "projects": [{"project_path": "/path", "commands": 6766, ...}],
  "tool_breakdown": [{"tool_name": "bash", "count": 13940, ...}]
}
```

To metrics-api per-project format:

```json
{
  "id": "project_Sprint_12",
  "developer": "e40057167",
  "project": "project",
  "sprint": "Sprint 12",
  "bash": 6766,
  "reads": 624,
  "edits": 624,
  "searches": 52,
  "time_saved_bash": 3857.14,
  "time_saved_total": 4284.52
}
```

## Configuration

Configuration file: `~/.config/daily-metrics/config.json`

The tool will automatically create a default config file on first run. You can edit it to customize:

```json
{
  "work_repos": [
    "/home/user/projects/work-repo-1",
    "/home/user/projects/work-repo-2"
  ],
  "private_repos": [
    "/home/user/projects/personal-project"
  ],
  "typing_speed_wpm": 50,
  "jira_user": "your.email@company.com",
  "git_email": "your.email@company.com"
}
```

**Config Fields:**

- `work_repos` - List of work/public repository paths
- `private_repos` - List of private repository paths
- `typing_speed_wpm` - Your average typing speed (default: 50 WPM)
- `jira_user` - Your Jira email for filtering issues
- `git_email` - Your Git email for filtering commits
- `schedule_b_anchor` - Date for Schedule B Friday calculation (YYYY-MM-DD)
- `holidays_2025` / `holidays_2026` - Holiday dates to exclude from working days

## Architecture

### Phase 2: API Integration

The tool uses a smart fallback architecture:

1. **API Mode (preferred)**: Queries metrics-api for team data
   - Centralized caching (5-minute TTL)
   - Consistent data across team
   - Reduced API load on Jira/GitHub
2. **Offline Mode (fallback)**: Direct CLI queries when API unavailable
   - Uses `jira`, `gh`, and `git` commands
   - Works without network/API access
   - Slower (no caching)

**Automatic Detection:**

```bash
# Will try API first, fall back to CLI if unavailable
daily-metrics

# Force offline mode (skip API check)
daily-metrics --force-offline
```

### Personal vs Team Metrics

Per [ADR-001](../../docs/adr/001-personal-vs-team-metrics.md):

- **Personal metrics** (OpenCode AI usage) stay local - never sync to API
- **Team metrics** (Jira, GitHub, Git) can use API or CLI
- Privacy-first: No individual surveillance

## Requirements

- Go 1.26+
- OpenCode installed (for `opencode` subcommand)
- **Optional**: metrics-api running on localhost:8080
- Jira CLI (`jira`) for Jira integration (fallback)
- GitHub CLI (`gh`) for GitHub integration (fallback)
- Git

## Examples

### Generate today's report and save to Obsidian

```bash
daily-metrics --output ~/SecondBrain/2.\ Area/Daily\ Notes/$(date +%Y-%m-%d).md
```

### Weekly summary

```bash
daily-metrics summary --start 2026-04-10 --end 2026-04-17 --all
```

### Track AI time savings

```bash
daily-metrics opencode --days 7
```

## Development

Run tests:

```bash
go test -v
```

Build:

```bash
go build
```

Install locally:

```bash
go install
```
