import hashlib
from datetime import datetime
from typing import Any

import numpy as np
import pandas as pd

from .config import HISTORICAL_SPRINTS, SprintConfig

# Simple in-memory cache with TTL
_SPRINT_CACHE = {}
_CACHE_TTL_SECONDS = 3600  # 1 hour


def calculate_focus_factor(historical_df: pd.DataFrame) -> float:
    """
    Derive the team's focus factor from historical sprint entries.

    For each past sprint, computes effective team capacity from the
    provided contributors list (length, holidays, PTOs, commitment).
    Then calculates:
        focus_factor = avg(completed points) / avg(effective capacity days)

    Returns 0.70 if no historical data exists or computed capacity is zero.
    """
    if historical_df.empty:
        print("⚠️ No historical data; defaulting focus factor to 0.70")
        return 0.70

    capacities = []  # effective days per sprint
    for sprint, row in historical_df.iterrows():
        members = row.get("contributors")
        if isinstance(members, list):
            cap = calculate_actual_capacity_from_team(members, row.get("sprint_length_days", 0), row.get("holidays", 0))
        else:
            # This should no longer occur if config always includes contributors
            contributors = int(row.get("contributors", 0))
            days = max(0, row.get("sprint_length_days", 0) - row.get("holidays", 0))
            cap = contributors * days
        capacities.append(cap)

    avg_capacity = float(np.mean(capacities))
    avg_completed = float(historical_df["completed"].mean())

    if avg_capacity <= 0:
        print("⚠️ Computed zero capacity; defaulting focus factor to 0.70")
        return 0.70

    focus_factor = avg_completed / avg_capacity
    print("📊 Focus Factor Calculation:")
    print(f"   Avg completed points : {avg_completed:.1f}")
    print(f"   Avg effective capacity days : {avg_capacity:.1f}")
    print(f"   Focus factor : {focus_factor:.3f}\n")

    return focus_factor


def calculate_actual_capacity_from_team(contributors: list[dict[str, Any]], sprint_length: int, holidays: int) -> float:
    """
    Compute total effective capacity (person-days) for a sprint:
        effective_days = (sprint_length - holidays - pto_days) * commitment_percentage
    Sums across all members in the provided list.
    """
    total_capacity = 0.0
    work_days = max(0, sprint_length - holidays)
    for contributor in contributors:
        commitment_pct = contributor.get("commitment", contributor.get("commitment_percentage", 1.0))
        pto = contributor.get("pto_days", 0)
        available_days = max(0, work_days - pto)
        total_capacity += available_days * commitment_pct
    return total_capacity


def add_completed_sprint_helper(
    sprint_name: str,
    committed: int,
    completed: int,
    carryover: int,
    sprint_length: int = 10,
    holidays: int = 0,
    contributors: list[dict[str, Any]] = None,
) -> dict[str, Any]:
    """
    Interactive helper to generate a HISTORICAL_SPRINTS entry.
    Prompts for contributors to compute capacity, then returns a dict
    including the full list for transparency.  Example usage:

        entry = add_completed_sprint_helper(
            "Sprint 14", 85, 78, 12,
            sprint_length=10, holidays=1,
            contributors=[ ... ]
        )
    """
    if not contributors:
        raise ValueError("Provide a list of contributors to compute capacity.")

    calculate_actual_capacity_from_team(contributors, sprint_length, holidays)
    sprint_data = {
        "sprint": sprint_name,
        "committed": committed,
        "completed": completed,
        "carryover": carryover,
        "sprint_length_days": sprint_length,
        "holidays": holidays,
        "contributors": contributors,
    }
    print("🔧 Add this dict to HISTORICAL_SPRINTS in config.py:")
    print(sprint_data)
    return sprint_data


def calculate_sprint_capacity(sprint_config: SprintConfig, focus_factor: float) -> dict[str, Any]:
    """
    Plan capacity for the upcoming sprint:
    - Builds a DataFrame of each member's effective days
    - Sums total capacity and multiplies by focus factor for recommended points
    Returns: { capacity_df, total_capacity_days, recommended_points, focus_factor, sprint_config }
    """
    team_capacity = []
    total_available_days = max(0, sprint_config.length_days - sprint_config.holidays_in_sprint)
    for contributor in sprint_config.contributors:
        personal_available_days = max(0, total_available_days - contributor.pto_days)
        effective_capacity_days = personal_available_days * contributor.commitment_percentage
        team_capacity.append(
            {
                "Name": contributor.name,
                "Commitment": f"{contributor.commitment_percentage:.0%}",
                "PTO Days": contributor.pto_days,
                "Effective Days": round(effective_capacity_days, 2),
            }
        )
    df = pd.DataFrame(team_capacity)
    total_team_capacity_days = float(df["Effective Days"].sum())
    recommended_story_points = total_team_capacity_days * focus_factor
    return {
        "capacity_df": df,
        "total_capacity_days": total_team_capacity_days,
        "recommended_points": recommended_story_points,
        "focus_factor": focus_factor,
        "sprint_config": sprint_config,
    }


def _get_cache_key(
    project_key: str, team_id: str, count: int, workflow_type: str, start_date: str, end_date: str
) -> str:
    """Generate cache key from request parameters."""
    key_parts = f"{project_key}:{team_id}:{count}:{workflow_type}:{start_date}:{end_date}"
    return hashlib.md5(key_parts.encode()).hexdigest()


def get_historical_data_from_jira(
    project_key: str,
    team_id: str = None,
    count: int = 5,
    workflow_type: str = "agile",
    start_date: str = None,
    end_date: str = None,
) -> pd.DataFrame:
    """
    Fetch historical sprint or Kanban data from Jira using JiraClient.
    Returns DataFrame compatible with existing analytics functions.

    Server-side caching: Results cached for 1 hour based on parameters.

    Args:
        project_key: Jira project key (e.g., 'PROJ')
        team_id: Optional team ID to filter by (e.g., '2490')
        count: Number of sprints/weeks to fetch
        workflow_type: 'agile' (sprints) or 'kanban' (time-based)
        start_date: Optional start date for Kanban mode (YYYY-MM-DD)
        end_date: Optional end date for Kanban mode (YYYY-MM-DD)
    """
    # Check cache first
    cache_key = _get_cache_key(project_key, team_id or "", count, workflow_type, start_date or "", end_date or "")

    if cache_key in _SPRINT_CACHE:
        cached_data, cached_time = _SPRINT_CACHE[cache_key]
        age_seconds = (datetime.now() - cached_time).total_seconds()
        if age_seconds < _CACHE_TTL_SECONDS:
            print(f"✅ Cache hit! Age: {age_seconds:.0f}s / {_CACHE_TTL_SECONDS}s")
            return cached_data
        else:
            print(f"⏰ Cache expired (age: {age_seconds:.0f}s)")
            del _SPRINT_CACHE[cache_key]

    from .jira_client import JiraClient

    try:
        client = JiraClient(project_key, team_id=team_id)

        if workflow_type == "kanban":
            historical_data = client.get_kanban_metrics(weeks=count, start_date=start_date, end_date=end_date)
        else:
            historical_data = client.get_historical_sprint_data(count=count)

        if not historical_data:
            return pd.DataFrame()

        df = pd.DataFrame(historical_data)

        df["commitment_reliability_pct"] = (df["completed"] / df["committed"]).replace([np.inf, -np.inf], 0).fillna(
            0
        ) * 100
        df["contributor_count"] = df["contributors"].apply(
            lambda x: len(x) if isinstance(x, list) else (x if isinstance(x, (int, float)) else 0)
        )

        df["contributor_count"] = df["contributor_count"].replace(0, 1)
        df["points_per_contributor"] = (
            (df["completed"] / df["contributor_count"]).replace([np.inf, -np.inf], 0).fillna(0)
        )

        # Cache the result
        _SPRINT_CACHE[cache_key] = (df, datetime.now())
        print(f"💾 Cached sprint data (key: {cache_key[:8]}...)")

        return df
    except Exception as e:
        print(f"Warning: Failed to fetch Jira data: {e}")
        return pd.DataFrame()


def get_historical_data() -> pd.DataFrame:
    """
    Load and process HISTORICAL_SPRINTS:
    - Sets sprint as index
    - Computes commitment reliability %
    - Derives contributors count from contributors list
    - Calculates points_per_contributor

    DEPRECATED: Use get_historical_data_from_jira() instead.
    This function remains for backward compatibility with YAML config.
    """
    if not HISTORICAL_SPRINTS:
        return pd.DataFrame()

    df = pd.DataFrame(HISTORICAL_SPRINTS)

    # 1. Calculate Reliability
    # Use fillna(0) to handle cases where committed is 0 or missing
    df["commitment_reliability_pct"] = (df["completed"] / df["committed"]).fillna(0) * 100

    # 2. Calculate Contributor Count
    # We create a NEW column 'contributor_count' instead of overwriting 'contributors'
    # so we don't lose the original data (useful if we want to list names later).
    df["contributor_count"] = df["contributors"].apply(
        lambda x: len(x) if isinstance(x, list) else (x if isinstance(x, (int, float)) else 0)
    )

    # 3. Calculate Efficiency
    # Handle Division by Zero (Infinity) and Missing Values (NaN) explicitly
    # replacing them with 0 ensures the JSON sent to the frontend is valid.
    df["points_per_contributor"] = (df["completed"] / df["contributor_count"]).replace([np.inf, -np.inf], 0).fillna(0)

    return df


def get_dashboard_data_from_df(df: pd.DataFrame) -> dict[str, Any]:
    """
    Prepare data structures for frontend Chart.js from a DataFrame.
    """
    if df.empty:
        return {}

    avg_completed = float(df["completed"].mean()) if not df.empty else 0.0

    return {
        "labels": df["sprint"].tolist(),
        "workload": {
            "committed": df["committed"].tolist(),
            "completed": df["completed"].tolist(),
            "carryover": df["carryover"].tolist(),
            "average_completed": avg_completed,
        },
        "performance": {
            "contributors": df["contributor_count"].tolist(),
            "points_per_contributor": df["points_per_contributor"].round(1).tolist(),
            "reliability": df["commitment_reliability_pct"].round(1).tolist(),
        },
    }


def get_dashboard_data() -> dict[str, Any]:
    """
    Prepare data structures for frontend Chart.js.

    DEPRECATED: Use get_dashboard_data_from_df() instead.
    """
    df = get_historical_data()
    return get_dashboard_data_from_df(df)


def get_latest_sprint_metrics_summary_from_df(df: pd.DataFrame) -> dict[str, Any]:
    """
    Get high-level aggregated metrics for the last sprint from a DataFrame.
    """
    if df.empty:
        return {
            "sprint_name": "No Data",
            "committed_points": 0,
            "completed_points": 0,
            "carry_over": 0,
            "commitment_reliability": 0,
        }

    latest = df.iloc[-1]

    return {
        "sprint_name": latest["sprint"],
        "committed_points": int(latest["committed"]),
        "completed_points": int(latest["completed"]),
        "carry_over": int(latest["carryover"]),
        "commitment_reliability": round(latest["commitment_reliability_pct"], 1),
    }


def get_latest_sprint_metrics_summary() -> dict[str, Any]:
    """
    Get high-level aggregated metrics for the last sprint.

    DEPRECATED: Use get_latest_sprint_metrics_summary_from_df() instead.
    """
    df = get_historical_data()
    return get_latest_sprint_metrics_summary_from_df(df)


# TODO - update to get metric averages
def get_avg_sprint_metrics_summary() -> dict[str, Any]:
    """
    Get high-level aggregated metrics for the sprint averages.
    """
    df = get_historical_data()

    if df.empty:
        return {
            "sprint_name": "No Data",
            "committed_points": 0,
            "completed_points": 0,
            "carry_over": 0,
            "commitment_reliability": 0,
        }

    # Get the most recent sprint's data for the summary cards
    latest = df.iloc[-1]

    return {
        "sprint_name": latest["sprint"],
        "committed_points": int(latest["committed"]),
        "completed_points": int(latest["completed"]),
        "carry_over": int(latest["carryover"]),
        "commitment_reliability": round(latest["commitment_reliability_pct"], 1),
    }


def calculate_velocity_trend(velocities: list[float], window: int = 5) -> float | None:
    """
    Calculates average sprint-to-sprint percentage change in velocity
    over a rolling window.

    Returns:
        Average percentage change (e.g. 0.042 = +4.2%)
        None if not enough data
    """
    if len(velocities) < 2:
        return None

    # Take the most recent window
    recent = velocities[-window:]

    percent_changes = []

    for prev, curr in zip(recent[:-1], recent[1:]):
        if prev == 0:
            continue  # avoid divide-by-zero lies
        percent_changes.append((curr - prev) / prev)

    if not percent_changes:
        return None

    return (sum(percent_changes) / len(percent_changes)) * 100


def print_capacity_report(capacity_result: dict[str, Any]) -> None:
    """Print formatted capacity planning report."""
    config = capacity_result["sprint_config"]
    df = capacity_result["capacity_df"]
    focus_factor = capacity_result["focus_factor"]

    print("=" * 50)
    print("          SPRINT CAPACITY PLAN")
    print("=" * 50)
    print(f"Sprint Length: {config.length_days} days")
    print(f"Team Holidays: {config.holidays_in_sprint} days")
    print(f"Focus Factor: {focus_factor:.3f} (calculated from historical data)")
    print()

    print(df.to_string(index=False))
    print("-" * 50)
    print(f"Total Effective Capacity: {capacity_result['total_capacity_days']:.2f} person-days")
    print(f"✅ Recommended Story Points: {capacity_result['recommended_points']:.0f}")
    print("=" * 50)


def get_team_metrics(df: pd.DataFrame) -> dict[str, float]:
    """Calculate key team performance metrics."""
    return {
        "avg_velocity": df["completed"].mean(),
        "avg_reliability": round(df["commitment_reliability_pct"].mean(), 2),
        "avg_points_per_contributor": round(df["points_per_contributor"].mean(), 2),
        "velocity_trend": round(df["completed"].pct_change().mean() * 100, 2),
        "reliability_trend": round(df["commitment_reliability_pct"].pct_change().mean() * 100, 2),
    }
