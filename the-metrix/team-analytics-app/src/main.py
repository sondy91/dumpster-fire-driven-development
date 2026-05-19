from contextlib import asynccontextmanager
from pathlib import Path

from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

from . import sprint_utils
from .api_client import MetricsAPIClient


# Lifespan instead of deprecated @app.on_event
@asynccontextmanager
async def lifespan(app: FastAPI):
    api_client = MetricsAPIClient()
    app.state.api_client = api_client
    app.state.api_available = api_client.is_available()

    try:
        data = sprint_utils.get_historical_data()
        app.state.ready = True if data else False
    except Exception:
        app.state.ready = False
    yield

    api_client.close()
    app.state.ready = False
    app.state.api_available = False


app = FastAPI(lifespan=lifespan)

BASE_DIR = Path(__file__).parent
app.mount("/static", StaticFiles(directory=str(BASE_DIR / "static")), name="static")
templates = Jinja2Templates(directory=str(BASE_DIR / "templates"))


@app.get("/", response_class=HTMLResponse)
def root(request: Request):
    """Root redirects to insights (main landing page)."""
    from fastapi.responses import RedirectResponse

    return RedirectResponse(url="/insights")


@app.get("/sprint", response_class=HTMLResponse)
def sprint_dashboard(
    request: Request,
    project: str = "PROJ",
    team_id: str = "2490",
    workflow: str = "agile",
    count: int = 5,
    start_date: str = None,
    end_date: str = None,
    use_cache: bool = False,
    skeleton: bool = True,
):
    """
    Sprint/Kanban analytics dashboard.

    Query params:
    - project: Jira project key (default: PROJ)
    - team_id: Jira team ID to filter by (default: 2490)
    - workflow: 'agile' or 'kanban' (default: agile)
    - count: Number of sprints/weeks to show (default: 5 for agile, 4 for kanban)
    - start_date: Optional start date for Kanban mode (YYYY-MM-DD)
    - end_date: Optional end date for Kanban mode (YYYY-MM-DD)
    - use_cache: If true, skip Jira fetch (frontend will handle caching)
    - skeleton: If true, render skeleton immediately and load data via HTMX (default: true)
    """
    if workflow == "kanban" and count == 5:
        count = 4

    # Render skeleton template immediately for fast page load
    if skeleton:
        page_title = "Kanban Analytics" if workflow == "kanban" else "Sprint Analytics"
        return templates.TemplateResponse(
            "sprint_skeleton.html",
            {
                "request": request,
                "project": project,
                "team_id": team_id,
                "workflow": workflow,
                "count": count,
                "start_date": start_date or "",
                "end_date": end_date or "",
                "page_title": page_title,
            },
        )

    if use_cache:
        hd = sprint_utils.get_historical_data()
    else:
        try:
            hd = sprint_utils.get_historical_data_from_jira(
                project, team_id=team_id, count=count, workflow_type=workflow, start_date=start_date, end_date=end_date
            )

            if hd.empty:
                hd = sprint_utils.get_historical_data()
        except Exception as e:
            print(f"Warning: Jira fetch failed, using config: {e}")
            hd = sprint_utils.get_historical_data()

    if hd.empty:
        latest_sprint_metrics = {
            "sprint_name": "No Data",
            "committed_points": 0,
            "completed_points": 0,
            "carry_over": 0,
            "commitment_reliability": 0,
        }
        chart_data = {}
        velocity_trend = 0
        reliability_trend = 0
        focus_factor = 0
        team_metrics = {
            "avg_velocity": 0,
            "focus_factor": 0,
            "commitment_reliability": 0,
        }
    else:
        latest_sprint_metrics = sprint_utils.get_latest_sprint_metrics_summary_from_df(hd)
        chart_data = sprint_utils.get_dashboard_data_from_df(hd)
        
        focus_factor = sprint_utils.calculate_focus_factor(hd)
        team_metrics = sprint_utils.get_team_metrics(hd)
        velocity_trend = sprint_utils.calculate_velocity_trend(chart_data["workload"]["completed"])
        reliability_trend = sprint_utils.calculate_velocity_trend(chart_data["performance"]["reliability"])

    api_status = {
        "available": getattr(request.app.state, "api_available", False),
        "mode": "jira" if not hd.empty else "config",
    }

    page_title = "Kanban Analytics" if workflow == "kanban" else "Sprint Analytics"

    # Use content-only template when loading via HTMX (no layout duplication)
    template_name = "sprint_content.html" if not skeleton else "sprint.html"

    return templates.TemplateResponse(
        template_name,
        {
            "request": request,
            "latest_sprint_metrics": latest_sprint_metrics,
            "chart_data": chart_data,
            "focus_factor": round(focus_factor, 3),
            "team_metrics": team_metrics,
            "velocity_trend": f"{round(velocity_trend, 2):+}",
            "page_title": page_title,
            "workflow_type": workflow,
            "reliability_trend": f"{round(reliability_trend, 2):+}",
            "api_status": api_status,
        },
    )


@app.get("/healthz")
def healthz():
    return {"ok": True}


@app.get("/readyz")
def readyz():
    return ({"ready": True}, 200) if getattr(app.state, "ready", False) else ({"ready": False}, 503)


@app.get("/api/status")
def api_status(request: Request):
    """Check metrics-api availability and connection status."""
    api_available = getattr(request.app.state, "api_available", False)
    api_client = getattr(request.app.state, "api_client", None)

    status = {
        "api_available": api_available,
        "api_url": api_client.base_url if api_client else None,
        "mode": "api" if api_available else "config",
    }

    if api_client and api_available:
        try:
            health = api_client.health()
            status["api_health"] = health
        except Exception as e:
            status["api_health"] = {"error": str(e)}

    return status


@app.get("/metrics", response_class=HTMLResponse)
async def live_metrics(request: Request, start_date: str = None, end_date: str = None):
    """Show live metrics page (data loaded via API after page loads)."""
    from datetime import date, timedelta

    api_available = getattr(request.app.state, "api_available", False)

    # Default to last 7 days
    if not end_date:
        end_date = date.today().isoformat()
    if not start_date:
        start_date = (date.today() - timedelta(days=7)).isoformat()

    # Return page immediately with skeleton loading
    return templates.TemplateResponse(
        "metrics.html",
        {
            "request": request,
            "metrics": {
                "api_available": api_available,
                "start_date": start_date,
                "end_date": end_date,
                "loading": True,  # Signal to show skeletons
            },
        },
    )


@app.get("/api/metrics/all")
async def get_all_metrics(request: Request, start_date: str = None, end_date: str = None):
    """API endpoint for fetching all metrics data."""
    from datetime import date, datetime, timedelta

    api_available = getattr(request.app.state, "api_available", False)
    api_client = getattr(request.app.state, "api_client", None)

    # Default to last 7 days
    if not end_date:
        end_date = date.today().isoformat()
    if not start_date:
        start_date = (date.today() - timedelta(days=7)).isoformat()

    # Validate date format
    validation_error = None
    try:
        start_dt = datetime.strptime(start_date, "%Y-%m-%d")
        end_dt = datetime.strptime(end_date, "%Y-%m-%d")
        if start_dt > end_dt:
            validation_error = "Start date must be before end date"
        elif end_dt > datetime.now():
            validation_error = "End date cannot be in the future"
    except ValueError:
        validation_error = "Invalid date format (use YYYY-MM-DD)"

    metrics_data = {
        "api_available": api_available,
        "start_date": start_date,
        "end_date": end_date,
        "jira": None,
        "github": None,
        "git": None,
        "errors": [],
    }

    if validation_error:
        metrics_data["errors"].append(validation_error)
    elif api_client and api_available:
        # Fetch all metrics in parallel for ~3x faster loading
        results = await api_client.get_all_metrics_parallel(start_date, end_date)
        metrics_data["jira"] = results.get("jira")
        metrics_data["github"] = results.get("github")
        metrics_data["git"] = results.get("git")
        metrics_data["errors"] = results.get("errors", [])

    return metrics_data


@app.get("/metrics-old", response_class=HTMLResponse)
async def live_metrics_old(request: Request, start_date: str = None, end_date: str = None):
    """Old server-rendered metrics (kept for reference)."""
    from datetime import date, datetime, timedelta

    api_available = getattr(request.app.state, "api_available", False)
    api_client = getattr(request.app.state, "api_client", None)

    # Default to last 7 days
    if not end_date:
        end_date = date.today().isoformat()
    if not start_date:
        start_date = (date.today() - timedelta(days=7)).isoformat()

    # Validate date format
    validation_error = None
    try:
        start_dt = datetime.strptime(start_date, "%Y-%m-%d")
        end_dt = datetime.strptime(end_date, "%Y-%m-%d")
        if start_dt > end_dt:
            validation_error = "Start date must be before end date"
        elif end_dt > datetime.now():
            validation_error = "End date cannot be in the future"
    except ValueError:
        validation_error = "Invalid date format (use YYYY-MM-DD)"

    metrics_data = {
        "api_available": api_available,
        "start_date": start_date,
        "end_date": end_date,
        "jira": None,
        "github": None,
        "git": None,
        "errors": [],
    }

    if validation_error:
        metrics_data["errors"].append(validation_error)
    elif api_client and api_available:
        # Fetch all metrics in parallel for ~3x faster loading
        results = await api_client.get_all_metrics_parallel(start_date, end_date)
        metrics_data["jira"] = results.get("jira")
        metrics_data["github"] = results.get("github")
        metrics_data["git"] = results.get("git")
        metrics_data["errors"] = results.get("errors", [])

    return templates.TemplateResponse(
        "metrics.html",
        {
            "request": request,
            "metrics": metrics_data,
        },
    )


@app.get("/dora", response_class=HTMLResponse)
async def dora_metrics(request: Request, start_date: str = None, end_date: str = None):
    """Show DORA metrics page (data loaded via API after page loads)."""
    from datetime import date, timedelta

    api_available = getattr(request.app.state, "api_available", False)

    if not end_date:
        end_date = date.today().isoformat()
    if not start_date:
        start_date = (date.today() - timedelta(days=30)).isoformat()

    # Return page immediately with skeleton loading
    return templates.TemplateResponse(
        "dora.html",
        {
            "request": request,
            "start_date": start_date,
            "end_date": end_date,
            "api_available": api_available,
            "loading": True,  # Signal to show skeletons
        },
    )


@app.get("/api/dora/metrics")
async def get_dora_metrics(request: Request, start_date: str = None, end_date: str = None):
    """API endpoint for DORA metrics calculation."""
    from datetime import date, datetime, timedelta

    api_available = getattr(request.app.state, "api_available", False)
    api_client = getattr(request.app.state, "api_client", None)

    if not end_date:
        end_date = date.today().isoformat()
    if not start_date:
        start_date = (date.today() - timedelta(days=30)).isoformat()

    # Default values when no data available (not demo mode)
    dora_data = {
        "df_value": "—",
        "df_unit": "",
        "df_sub": "No data available",
        "df_class": "none",
        "df_badge": "N/A",
        "lt_value": "—",
        "lt_unit": "",
        "lt_sub": "No data available",
        "lt_class": "none",
        "lt_badge": "N/A",
        "cfr_value": "—",
        "cfr_unit": "",
        "cfr_sub": "No data available",
        "cfr_class": "none",
        "cfr_badge": "N/A",
        "mttr_value": "—",
        "mttr_unit": "",
        "mttr_sub": "No data available",
        "mttr_class": "none",
        "mttr_badge": "N/A",
    }

    if api_client and api_available:
        try:
            results = await api_client.get_all_metrics_parallel(start_date, end_date)
            github_metrics = results.get("github", {})

            start_dt = datetime.strptime(start_date, "%Y-%m-%d")
            end_dt = datetime.strptime(end_date, "%Y-%m-%d")
            days = (end_dt - start_dt).days + 1
            weeks = days / 7

            merged_count = github_metrics.get("merged_prs", 0)
            per_week = merged_count / weeks if weeks > 0 else 0

            if per_week >= 7:
                dora_data["df_class"] = "elite"
                dora_data["df_badge"] = "Elite"
            elif per_week >= 1:
                dora_data["df_class"] = "high"
                dora_data["df_badge"] = "High"
            elif per_week >= 0.25:
                dora_data["df_class"] = "med"
                dora_data["df_badge"] = "Medium"
            else:
                dora_data["df_class"] = "low"
                dora_data["df_badge"] = "Low"

            dora_data["df_value"] = f"{per_week:.1f}"
            dora_data["df_unit"] = "/week"
            dora_data["df_sub"] = f"{merged_count} PRs merged in {days} days"

            prs_data = github_metrics.get("prs", [])
            merged_prs = [
                pr for pr in prs_data if pr.get("state") == "merged" and pr.get("created_at") and pr.get("merged_at")
            ]

            if merged_prs:
                lead_times_hours = []
                for pr in merged_prs:
                    created = datetime.fromisoformat(pr["created_at"].replace("Z", "+00:00"))
                    merged_at = datetime.fromisoformat(pr["merged_at"].replace("Z", "+00:00"))
                    hours = (merged_at - created).total_seconds() / 3600
                    lead_times_hours.append(hours)

                avg_hours = sum(lead_times_hours) / len(lead_times_hours)

                if avg_hours < 1:
                    dora_data["lt_class"] = "elite"
                    dora_data["lt_badge"] = "Elite"
                elif avg_hours < 24:
                    dora_data["lt_class"] = "high"
                    dora_data["lt_badge"] = "High"
                elif avg_hours < 168:
                    dora_data["lt_class"] = "med"
                    dora_data["lt_badge"] = "Medium"
                else:
                    dora_data["lt_class"] = "low"
                    dora_data["lt_badge"] = "Low"

                if avg_hours < 24:
                    dora_data["lt_value"] = f"{avg_hours:.1f}"
                    dora_data["lt_unit"] = "hours"
                else:
                    dora_data["lt_value"] = f"{avg_hours / 24:.1f}"
                    dora_data["lt_unit"] = "days"

                dora_data["lt_sub"] = f"Average from {len(merged_prs)} merged PRs"

            # Calculate Change Failure Rate (CFR)
            # Use GitHub labels to identify bug fixes (more reliable than title parsing)
            bug_labels = {"bug", "bugfix", "hotfix", "fix", "revert", "patch"}
            failed_changes = [
                pr
                for pr in merged_prs
                if any(label.get("name", "").lower() in bug_labels for label in pr.get("labels", []))
            ]

            if merged_count > 0:
                cfr_percentage = (len(failed_changes) / merged_count) * 100

                if cfr_percentage <= 5:
                    dora_data["cfr_class"] = "elite"
                    dora_data["cfr_badge"] = "Elite"
                elif cfr_percentage <= 10:
                    dora_data["cfr_class"] = "high"
                    dora_data["cfr_badge"] = "High"
                elif cfr_percentage <= 15:
                    dora_data["cfr_class"] = "med"
                    dora_data["cfr_badge"] = "Medium"
                else:
                    dora_data["cfr_class"] = "low"
                    dora_data["cfr_badge"] = "Low"

                dora_data["cfr_value"] = f"{cfr_percentage:.1f}"
                dora_data["cfr_unit"] = "%"
                dora_data["cfr_sub"] = f"{len(failed_changes)} fixes out of {merged_count} changes"

            # Calculate Mean Time to Restore (MTTR)
            # Approximate by measuring time to merge fix PRs
            if failed_changes:
                restore_times_hours = []
                for pr in failed_changes:
                    created = datetime.fromisoformat(pr["created_at"].replace("Z", "+00:00"))
                    merged_at = datetime.fromisoformat(pr["merged_at"].replace("Z", "+00:00"))
                    hours = (merged_at - created).total_seconds() / 3600
                    restore_times_hours.append(hours)

                avg_restore_hours = sum(restore_times_hours) / len(restore_times_hours)

                if avg_restore_hours < 1:
                    dora_data["mttr_class"] = "elite"
                    dora_data["mttr_badge"] = "Elite"
                elif avg_restore_hours < 24:
                    dora_data["mttr_class"] = "high"
                    dora_data["mttr_badge"] = "High"
                elif avg_restore_hours < 168:
                    dora_data["mttr_class"] = "med"
                    dora_data["mttr_badge"] = "Medium"
                else:
                    dora_data["mttr_class"] = "low"
                    dora_data["mttr_badge"] = "Low"

                if avg_restore_hours < 24:
                    dora_data["mttr_value"] = f"{avg_restore_hours:.1f}"
                    dora_data["mttr_unit"] = "hours"
                else:
                    dora_data["mttr_value"] = f"{avg_restore_hours / 24:.1f}"
                    dora_data["mttr_unit"] = "days"

                dora_data["mttr_sub"] = f"Average from {len(failed_changes)} fix PRs"

        except Exception:
            pass

    return dora_data


@app.get("/survey", response_class=HTMLResponse)
def survey(request: Request):
    """Team pulse survey page."""
    return templates.TemplateResponse(
        "survey.html",
        {
            "request": request,
        },
    )


@app.get("/opencode", response_class=HTMLResponse)
def opencode(request: Request):
    """OpenCode AI stats tracking page."""
    return templates.TemplateResponse(
        "opencode.html",
        {
            "request": request,
        },
    )


@app.get("/insights", response_class=HTMLResponse)
def insights(request: Request):
    """Main insights landing page with aggregated metrics (with 5-minute cache)."""
    from datetime import datetime, timedelta

    # Simple in-memory cache with 5-minute TTL
    cache_key = "insights_data"
    cache_ttl = timedelta(minutes=5)

    # Check if cached data exists and is still valid
    cached_data = getattr(request.app.state, f"{cache_key}_data", None)
    cached_time = getattr(request.app.state, f"{cache_key}_time", None)

    if cached_data and cached_time and (datetime.now() - cached_time) < cache_ttl:
        print(f"✓ Using cached insights data (age: {(datetime.now() - cached_time).seconds}s)")
        return templates.TemplateResponse("insights.html", cached_data)

    # Cache miss or expired - fetch fresh data
    print("⟳ Fetching fresh insights data from Jira...")
    hd = sprint_utils.get_historical_data()
    focus_factor = sprint_utils.calculate_focus_factor(hd)
    team_metrics = sprint_utils.get_team_metrics(hd)

    template_data = {
        "request": request,
        "team_metrics": team_metrics,
        "focus_factor": round(focus_factor, 3),
    }

    # Store in cache
    setattr(request.app.state, f"{cache_key}_data", template_data)
    setattr(request.app.state, f"{cache_key}_time", datetime.now())

    return templates.TemplateResponse("insights.html", template_data)


@app.get("/settings", response_class=HTMLResponse)
def settings(request: Request):
    """User settings page for configuring workflow type and project."""
    return templates.TemplateResponse(
        "settings.html",
        {
            "request": request,
        },
    )


@app.get("/api/sprint-data")
def get_sprint_data(project: str = "PROJ", count: int = 5, workflow: str = None):
    """
    API endpoint to fetch sprint data from Jira.
    Returns JSON for frontend caching.
    Auto-detects workflow type if not specified.
    """
    from datetime import datetime

    from fastapi.responses import JSONResponse

    try:
        # Try agile mode first if not specified
        if workflow is None or workflow == "agile":
            df = sprint_utils.get_historical_data_from_jira(project, count=count, workflow_type="agile")

            # If no sprint data, auto-switch to Kanban
            if df.empty:
                print(f"No sprint data found for {project}, trying Kanban mode...")
                df = sprint_utils.get_historical_data_from_jira(project, count=count, workflow_type="kanban")
                workflow = "kanban"
            else:
                workflow = "agile"
        else:
            # Workflow explicitly specified
            df = sprint_utils.get_historical_data_from_jira(project, count=count, workflow_type=workflow)

        if df.empty:
            return JSONResponse(content={"error": "No data found", "data": [], "workflow": workflow}, status_code=404)

        data = df.to_dict(orient="records")

        return JSONResponse(
            content={
                "project": project,
                "count": count,
                "workflow": workflow,  # Return detected workflow
                "timestamp": datetime.now().isoformat(),
                "data": data,
            }
        )
    except Exception as e:
        return JSONResponse(content={"error": str(e), "data": [], "workflow": workflow}, status_code=500)


@app.get("/api/projects")
def get_projects():
    """
    API endpoint to fetch all Jira projects user has access to.
    Returns JSON with project keys and names.
    """
    from fastapi.responses import JSONResponse

    from .jira_client import JiraClient

    try:
        projects = JiraClient.get_all_projects()

        return JSONResponse(
            content={
                "projects": projects,
            }
        )
    except Exception as e:
        return JSONResponse(content={"error": str(e), "projects": []}, status_code=500)


@app.get("/api/teams")
def get_teams(project: str = "PROJ"):
    """
    API endpoint to fetch unique teams from a Jira project.
    Returns JSON with team IDs and names.
    """
    from fastapi.responses import JSONResponse

    from .jira_client import JiraClient

    try:
        client = JiraClient(project)
        teams = client.get_unique_teams(limit=100)

        return JSONResponse(
            content={
                "project": project,
                "teams": teams,
            }
        )
    except Exception as e:
        return JSONResponse(content={"error": str(e), "teams": []}, status_code=500)


@app.get("/feedback", response_class=HTMLResponse)
def feedback_admin(request: Request):
    """Admin view of feedback submissions."""
    return templates.TemplateResponse("feedback_admin.html", {"request": request})


# OpenCode proxy endpoints
@app.get("/api/opencode/activity")
def get_opencode_activity(
    request: Request,
    start_date: str = None,
    end_date: str = None,
    developer: str = None,
    project_name: str = None,
    limit: int = 1000,
):
    """Proxy to Go API for OpenCode activity data."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        # Build query params
        params = {"limit": limit}
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        if developer:
            params["developer"] = developer
        if project_name:
            params["project_name"] = project_name

        response = api_client.client.get("/api/v1/opencode/activity", params=params)
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to fetch activity: {e!s}"}, status_code=500)


@app.get("/api/opencode/sessions")
def get_opencode_sessions(
    request: Request,
    developer: str = None,
    project: str = None,
    sprint: str = None,
    limit: int = 1000,
):
    """Proxy to Go API for OpenCode sessions (legacy manual entries)."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        params = {"limit": limit}
        if developer:
            params["developer"] = developer
        if project:
            params["project"] = project
        if sprint:
            params["sprint"] = sprint

        response = api_client.client.get("/api/v1/opencode/sessions", params=params)
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to fetch sessions: {e!s}"}, status_code=500)


@app.delete("/api/opencode/sessions")
def delete_opencode_session(request: Request, id: str):
    """Proxy to Go API for deleting OpenCode session."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    if not id:
        return JSONResponse(content={"error": "Session ID is required"}, status_code=400)

    try:
        response = api_client.client.delete("/api/v1/opencode/sessions", params={"id": id})
        response.raise_for_status()
        return JSONResponse(content={"message": "Session deleted successfully"})
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to delete session: {e!s}"}, status_code=500)


@app.post("/api/opencode/submit")
def submit_opencode_session(request: Request, session_data: dict):
    """Proxy to Go API for submitting OpenCode session (manual entry from UI)."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        response = api_client.client.post("/api/v1/opencode/submit", json=session_data)
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to submit session: {e!s}"}, status_code=500)


# GitHub metrics proxy endpoint
@app.get("/api/metrics/github")
def get_github_metrics(
    request: Request,
    start_date: str,
    end_date: str,
):
    """Proxy to Go API for GitHub metrics."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        response = api_client.client.get(
            "/api/v1/metrics/github", params={"start_date": start_date, "end_date": end_date}
        )
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to fetch GitHub metrics: {e!s}"}, status_code=500)


# Survey proxy endpoints
@app.get("/api/survey/responses")
def get_survey_responses(
    request: Request,
    sprint: str = None,
    limit: int = 1000,
):
    """Proxy to Go API for survey responses."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        params = {"limit": limit}
        if sprint:
            params["sprint"] = sprint

        response = api_client.client.get("/api/v1/survey/responses", params=params)
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to fetch survey responses: {e!s}"}, status_code=500)


@app.post("/api/survey/submit")
def submit_survey_response(request: Request, survey_data: dict):
    """Proxy to Go API for submitting survey response."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        response = api_client.client.post("/api/v1/survey/submit", json=survey_data)
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to submit survey: {e!s}"}, status_code=500)


@app.delete("/api/survey/responses")
def delete_all_survey_responses(request: Request):
    """Proxy to Go API for deleting all survey responses."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        response = api_client.client.delete("/api/v1/survey/responses")
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to delete survey responses: {e!s}"}, status_code=500)


# Feedback proxy endpoints
@app.get("/api/feedback")
def get_feedback(
    request: Request,
    page: str = None,
    kind: str = None,
    limit: int = 100,
):
    """Proxy to Go API for feedback submissions."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        params = {"limit": limit}
        if page:
            params["page"] = page
        if kind:
            params["kind"] = kind

        response = api_client.client.get("/api/v1/feedback", params=params)
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to fetch feedback: {e!s}"}, status_code=500)


@app.post("/api/feedback")
def submit_feedback(request: Request, feedback_data: dict):
    """Proxy to Go API for submitting feedback."""
    from fastapi.responses import JSONResponse

    api_client = getattr(request.app.state, "api_client", None)
    if not api_client:
        return JSONResponse(content={"error": "Metrics API client not available"}, status_code=503)

    try:
        response = api_client.client.post("/api/v1/feedback", json=feedback_data)
        response.raise_for_status()
        return response.json()
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to submit feedback: {e!s}"}, status_code=500)


# Component test pages (development only)
@app.get("/test/components.html", response_class=HTMLResponse)
def component_tests(request: Request):
    """Manual testing page for web components."""
    return templates.TemplateResponse("test/components.html", {"request": request})


@app.get("/test/loading.html", response_class=HTMLResponse)
def loading_animations(request: Request):
    """Preview different loading animations."""
    return templates.TemplateResponse("test/loading.html", {"request": request})


# Cache management endpoint
@app.post("/api/cache/clear")
def clear_cache(request: Request):
    """Clear the metrics cache."""
    from fastapi.responses import JSONResponse

    from src import api_client

    cleared_count = api_client.clear_metrics_cache()
    return JSONResponse(
        content={"success": True, "message": f"Cleared {cleared_count} cached entries"}, status_code=200
    )
