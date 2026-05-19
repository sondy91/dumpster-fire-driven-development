# Team Analytics Web App

Real-time team productivity dashboard providing sprint metrics, velocity trends, and workload analysis for engineering teams.

![Team Analytics Dashboard](./team-analytics-homepage.png)

## Purpose

Visualizes team performance using Jira data to track:

- Sprint velocity and completion rates
- Focus factor (planned vs unplanned work)
- Individual contributor metrics
- Workload distribution and reliability trends

Helps the team make data-driven decisions about sprint planning and capacity.

## Tech Stack

- **Backend**: FastAPI (Python 3.13+)
- **Frontend**: HTMX + Tailwind CSS
- **Data Sources**: metrics-api (with fallback to config.py)
- **HTTP Client**: httpx (modern, async-first)
- **Charts**: Chart.js via CDN
- **Deployment**: Azure App Service (containerized)

## Local Development

### Prerequisites

- Python 3.13+
- [uv](https://github.com/astral-sh/uv) package manager
- **Optional**: metrics-api running on localhost:8080 (falls back to config.py if unavailable)

### Setup

```bash
cd src/team-analytics-app

# Install dependencies
uv sync

# Run development server
uv run uvicorn src.main:app --reload --port 8000
```

Access at: `http://localhost:8000`

### Testing with metrics-api

To test API integration mode:

1. **Start metrics-api** (in separate terminal):

   ```bash
   cd /path/to/The\ Metrix/metrics-api
   go build ./cmd/server
   ./server
   ```

   Metrics API will start on <http://localhost:8080>

2. **Start team-analytics-app**:

   ```bash
   cd /path/to/The\ Metrix/team-analytics-app
   uv run uvicorn src.main:app --reload --port 8000
   ```

3. **Verify API mode**:

   ```bash
   curl http://localhost:8000/api/status
   ```

   Should show `"mode": "api"` and `"api_available": true`

If metrics-api is not running, the app will automatically fall back to config mode.

### Configuration

**User Settings** (via `/settings` page):

Users configure their preferences via the Settings page:

- **Jira Project Key**: Which Jira project to track (e.g., XAIP, CMS, PLAT)
- **Workflow Type**: Agile (sprints) or Kanban (time-based)
- **Date Ranges**: Custom date ranges for Kanban mode
- **Theme**: Dark or light mode
- **Demo Mode**: Toggle between live data and demo data

All settings are stored in browser localStorage and apply immediately.

**Sprint Capacity Planning** (optional):

For manual capacity planning, update `src/config.py`:

```python
CURRENT_SPRINT = SprintConfig(
    length_days=10,
    holidays_in_sprint=0,
    contributors=[
        Contributor("Name", commitment_percentage=1.0, pto_days=0),
        # ...
    ]
)
```

### Environment Variables

- `METRICS_API_URL` - Base URL for metrics-api (default: `http://localhost:8080`)

**Note**: Jira authentication is handled via `jira-cli` config (`~/.config/.jira/.config.yml`)

## Data Sources

### Jira Integration (Phase 2.5) - Primary

Sprint/Kanban analytics now pull live data from Jira via `jira-cli`:

- **Sprint Data**: Last N closed sprints with metrics
- **Kanban Data**: Completed issues grouped by week
- **Issue Details**: Story points, status, assignee
- **Metrics**: Velocity, reliability, focus factor

**Architecture:**

```
User Settings (localStorage)
    ↓
/sprint?project=XAIP&workflow=agile
    ↓
JiraClient (src/jira_client.py)
    ↓
jira-cli subprocess calls
    ↓
Jira REST API
```

**Performance:**

- Initial load: 10-30 seconds (fetches individual issue details)
- Cached: <1 second (data cached in localStorage for 1 hour)
- Background refresh available via `/api/sprint-data` endpoint

**Fallback:** If Jira fetch fails, falls back to `HISTORICAL_SPRINTS` in `config.py`

### metrics-api Integration (Phase 2)

DORA metrics and live metrics use metrics-api:

- **Jira metrics**: Issues, story points
- **GitHub metrics**: PRs, commits, deployment frequency
- **Git metrics**: LOC, working days

**Performance optimizations:**

- Parallel API fetching (Jira + GitHub + Git simultaneously)
- Response caching with 1hr TTL
- Async route handlers for non-blocking requests

Check API status:

```bash
curl http://localhost:8000/api/status
```

## Features

### Sprint/Kanban Analytics

- **Agile Mode**: Track last 5 sprints with velocity trends
- **Kanban Mode**: Track last 4 weeks (or custom date range) with throughput
- **Metrics**:
    - Velocity Trend (avg % change)
    - Commitment Reliability (completed vs committed)
    - Focus Factor (capacity-adjusted throughput)
    - Points per contributor
- **Charts**:
    - Workload analysis (committed, completed, carryover)
    - Performance trends (contributors, efficiency, reliability)

### DORA Metrics

- **Deployment Frequency**: PRs merged per week (Elite/High/Med/Low)
- **Lead Time for Changes**: Time from PR open to merge
- **Change Failure Rate**: Failed deploys % (placeholder)
- **Mean Time to Recover**: Recovery time for incidents (placeholder)

### Live Metrics Dashboard

Real-time stats from metrics-api:

- Jira issues (created, resolved, in progress)
- GitHub activity (PRs, commits, reviews)
- Git contributions (LOC, active days)

### Dev Pulse Survey

Team satisfaction tracking:

- Sprint-based survey responses
- Trend visualization over time
- Anonymous feedback collection

### AI Stats (OpenCode)

Track AI-assisted development:

- Time saved via OpenCode
- Session statistics
- Productivity metrics

## Deployment

Deployed via Helm chart to Azure App Service:

```bash
# From repo root
helm template team-analytics-app helm/team-analytics-app -f helm/team-analytics-app/values-dev.yaml

# Deploy via Argo CD (GitOps)
```

### Docker Build

```bash
docker build -t team-analytics-app .
docker run -p 8000:8000 team-analytics-app
```

## Integration

Works standalone but can integrate with:

- **metrics-api**: Use API to fetch Jira data instead of direct queries
- **daily-metrics**: CLI tool generates similar metrics for individual use

## Known Limitations

### Jira Integration Performance

**Issue:** Initial sprint data fetch is slow (10-30 seconds)

**Why:** `jira-cli` doesn't support bulk queries with custom fields. The app must call `jira issue view --raw` for each issue individually to get story points and sprint data.

**Mitigations:**

- **localStorage caching**: First load is slow, subsequent loads use cached data (1 hour TTL)
- **Background prefetch**: `/api/sprint-data` endpoint for async loading
- **Graceful fallback**: Falls back to config.py if Jira fetch fails

**Future improvements:**

- Direct Jira REST API calls (bypassing jira-cli)
- Server-side caching layer
- Background job to pre-fetch sprint data
- Progressive loading with skeleton UI

### Multi-User Deployment

**Current state:** Single-user tool using `jira-cli` config from `~/.config/.jira/`

**Limitation:** Different users need different Jira credentials and project access

**Options for multi-user:**

1. OAuth flow with per-user Jira tokens (requires backend auth)
2. Service account with read-only access (limited to single project)
3. SSO integration (future phase)

## Troubleshooting

### "No data available" or empty charts

**Cause**: Jira fetch failed or no sprint data found

**Solutions**:

1. Check Jira project key in Settings (`/settings`)
2. Verify `jira-cli` is configured: `jira me`
3. Test Jira access: `jira sprint list --project XAIP`
4. Check browser console for errors
5. Clear localStorage cache and reload

### Slow page load (Sprint Analytics)

**Expected behavior**: First load fetches from Jira (10-30 seconds)

**Solutions**:

1. Wait for initial load to complete
2. Subsequent loads use cache (<1 second)
3. Check "DEMO MODE" toggle if you need instant data for presentations

### Settings not persisting

**Cause**: localStorage is browser-specific

**Behavior**: Settings don't sync across browsers/devices (expected)

**Solution**: Use same browser or reconfigure settings

### Charts not rendering

- Check browser console for Chart.js errors
- Verify CDN is accessible
- Ensure data format matches chart expectations
- Try hard refresh (Ctrl+Shift+R)

## Related

- **Jira Board**: Configure your team's board URL in Jira settings
- **metrics-api**: Alternative data source (see `../metrics-api/`)
- **daily-metrics**: Individual productivity CLI tool (see `../daily-metrics/`)
