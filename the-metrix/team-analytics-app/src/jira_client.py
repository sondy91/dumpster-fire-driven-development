import json
import re
import subprocess
from datetime import date, datetime, timedelta
from typing import Any


class JiraClient:
    """Client for fetching Jira sprint and issue data via jira-cli subprocess calls."""

    def __init__(self, project_key: str = "PROJ", team_id: str = None):
        self.project_key = project_key
        self.team_id = team_id

    def _run_jira_command(self, args: list[str]) -> str:
        """Execute jira-cli command and return stdout."""
        try:
            result = subprocess.run(
                ["jira"] + args,
                capture_output=True,
                text=True,
                timeout=30,
            )
            if result.returncode != 0:
                raise RuntimeError(f"jira-cli error: {result.stderr}")
            return result.stdout
        except subprocess.TimeoutExpired:
            raise RuntimeError("jira-cli command timed out")
        except FileNotFoundError:
            raise RuntimeError("jira-cli not found. Is it installed?")

    def _parse_sprint_string(self, sprint_str: str) -> dict[str, Any] | None:
        """Parse Jira sprint string format into structured dict.

        Format: com.atlassian.greenhopper.service.sprint.Sprint@...[key=value,...]
        """
        if not sprint_str or sprint_str == "null":
            return None

        sprint_data = {}

        match = re.search(r"\[([^\]]+)\]", sprint_str)
        if not match:
            return None

        pairs = match.group(1).split(",")
        for pair in pairs:
            if "=" in pair:
                key, value = pair.split("=", 1)
                sprint_data[key.strip()] = value.strip()

        return {
            "id": sprint_data.get("id"),
            "name": sprint_data.get("name"),
            "state": sprint_data.get("state"),
            "start_date": sprint_data.get("startDate"),
            "end_date": sprint_data.get("endDate"),
            "complete_date": sprint_data.get("completeDate"),
            "goal": sprint_data.get("goal"),
        }

    def get_issues_with_sprints(self, limit: int = 100) -> list[dict]:
        """Get issue keys that have sprint data assigned."""
        jql = f"project = {self.project_key} AND Sprint is not EMPTY"

        if self.team_id:
            jql += f" AND Team = {self.team_id}"

        page_size = min(100, limit)

        output = self._run_jira_command(
            [
                "issue",
                "list",
                "--jql",
                jql,
                "--paginate",
                f"0:{page_size}",
                "--plain",
                "--columns",
                "key",
                "--no-headers",
            ]
        )

        issue_keys = []
        for line in output.strip().split("\n"):
            if line.strip():
                issue_keys.append(line.strip())

        return issue_keys[:limit]

    def get_issue_details(self, issue_key: str) -> dict | None:
        """Get full issue details including custom fields using --raw flag."""
        try:
            output = self._run_jira_command(["issue", "view", issue_key, "--raw"])

            data = json.loads(output)
            fields = data.get("fields", {})

            sprint_raw = fields.get("customfield_10105")
            sprints_data = []
            if sprint_raw and isinstance(sprint_raw, list):
                for sprint_str in sprint_raw:
                    parsed = self._parse_sprint_string(sprint_str)
                    if parsed:
                        sprints_data.append(parsed)

            story_points = fields.get("customfield_10106", 0.0)
            status = fields.get("status", {}).get("name", "")
            assignee_data = fields.get("assignee")
            assignee = assignee_data.get("displayName") if assignee_data else "Unassigned"

            team_data = fields.get("customfield_10842")
            team = None
            if team_data and isinstance(team_data, dict):
                team = {"id": str(team_data.get("id", "")), "name": team_data.get("name", "")}

            return {
                "key": issue_key,
                "sprints": sprints_data,
                "story_points": float(story_points) if story_points else 0.0,
                "status": status,
                "assignee": assignee,
                "team": team,
            }
        except Exception as e:
            print(f"Warning: Failed to get details for {issue_key}: {e}")
            return None

    @staticmethod
    def get_all_projects() -> list[dict]:
        """Get all Jira projects the user has access to.
        
        Returns list of {key, name} dicts.
        """
        try:
            result = subprocess.run(
                ["jira", "project", "list"],
                capture_output=True,
                text=True,
                timeout=10,
            )
            if result.returncode != 0:
                raise RuntimeError(f"jira-cli error: {result.stderr}")
            
            projects = []
            lines = result.stdout.strip().split("\n")
            
            # Skip header line
            for line in lines[1:]:
                if line.strip():
                    # Format: KEY\tNAME\tTYPE\tLEAD
                    parts = line.split("\t")
                    if len(parts) >= 2:
                        projects.append({
                            "key": parts[0].strip(),
                            "name": parts[1].strip()
                        })
            
            return projects
        except subprocess.TimeoutExpired:
            raise RuntimeError("jira-cli command timed out")
        except FileNotFoundError:
            raise RuntimeError("jira-cli not found. Is it installed?")
        except Exception as e:
            print(f"Warning: Failed to get projects: {e}")
            return []

    def get_unique_teams(self, limit: int = 100) -> list[dict]:
        """Get unique teams from project issues.

        Returns list of {id, name} dicts for team selection.
        """
        jql = f"project = {self.project_key}"

        output = self._run_jira_command(
            ["issue", "list", "--jql", jql, "--paginate", f"0:{limit}", "--plain", "--columns", "key", "--no-headers"]
        )

        issue_keys = []
        for line in output.strip().split("\n"):
            if line.strip():
                issue_keys.append(line.strip())

        teams_map = {}
        for issue_key in issue_keys[:30]:
            details = self.get_issue_details(issue_key)
            if details and details.get("team"):
                team = details["team"]
                team_id = team["id"]
                if team_id and team_id not in teams_map:
                    teams_map[team_id] = team

        teams = list(teams_map.values())
        teams.sort(key=lambda t: t.get("name", ""))

        return teams

    def get_unique_sprints(self, max_issues: int = 200) -> list[dict]:
        """Extract unique sprints from issues (workaround for Kanban boards).

        Note: This is slower as it fetches full details for each issue, but
        necessary since jira-cli doesn't support sprint list on Kanban boards.
        """
        issue_keys = self.get_issues_with_sprints(limit=max_issues)

        sprint_map = {}

        for issue_key in issue_keys[:50]:
            issue_details = self.get_issue_details(issue_key)
            if not issue_details:
                continue

            for sprint_data in issue_details.get("sprints", []):
                sprint_id = sprint_data.get("id")
                if sprint_id and sprint_id not in sprint_map:
                    sprint_map[sprint_id] = sprint_data

        sprints = list(sprint_map.values())
        sprints.sort(key=lambda s: s.get("start_date", ""), reverse=True)

        return sprints

    def get_sprint_issues(self, sprint_id: str) -> list[dict]:
        """Get all issues for a specific sprint with full details."""
        jql = f"project = {self.project_key} AND Sprint = {sprint_id}"

        if self.team_id:
            jql += f" AND Team = {self.team_id}"

        output = self._run_jira_command(["issue", "list", "--jql", jql, "--plain", "--columns", "key", "--no-headers"])

        issue_keys = []
        for line in output.strip().split("\n"):
            if line.strip():
                issue_keys.append(line.strip())

        issues = []
        for issue_key in issue_keys:
            details = self.get_issue_details(issue_key)
            if details:
                issues.append(
                    {
                        "key": details["key"],
                        "status": details["status"],
                        "story_points": details["story_points"],
                        "assignee": details["assignee"],
                    }
                )

        return issues

    def calculate_sprint_metrics(self, sprint_id: str) -> dict[str, Any]:
        """Calculate committed, completed, carryover for a sprint."""
        issues = self.get_sprint_issues(sprint_id)

        committed = 0.0
        completed = 0.0
        carryover = 0.0
        unique_assignees = set()

        done_statuses = {"Done", "Closed", "Resolved", "Cancelled"}

        for issue in issues:
            points = issue["story_points"]
            status = issue["status"]
            assignee = issue.get("assignee")

            committed += points

            if status in done_statuses:
                completed += points

            if assignee and assignee.strip():
                unique_assignees.add(assignee)

        carryover = committed - completed

        return {
            "committed": int(committed),
            "completed": int(completed),
            "carryover": int(carryover),
            "issue_count": len(issues),
            "contributor_count": len(unique_assignees) if unique_assignees else 1,
        }

    def get_historical_sprint_data(self, count: int = 5) -> list[dict[str, Any]]:
        """Get last N closed sprints with metrics."""
        sprints = self.get_unique_sprints(max_issues=100)

        closed_sprints = [s for s in sprints if s.get("state") == "CLOSED"]
        closed_sprints = closed_sprints[:count]

        historical_data = []

        for sprint in closed_sprints:
            sprint_id = sprint.get("id")
            if not sprint_id:
                continue

            metrics = self.calculate_sprint_metrics(sprint_id)

            start_date = sprint.get("start_date", "")
            end_date = sprint.get("end_date", "")

            sprint_length_days = 10
            if start_date and end_date:
                try:
                    start = datetime.fromisoformat(start_date.replace("Z", "+00:00"))
                    end = datetime.fromisoformat(end_date.replace("Z", "+00:00"))
                    sprint_length_days = (end - start).days
                except Exception:
                    pass

            historical_data.append(
                {
                    "sprint": sprint.get("name", f"Sprint {sprint_id}"),
                    "committed": metrics["committed"],
                    "completed": metrics["completed"],
                    "carryover": metrics["carryover"],
                    "sprint_length_days": sprint_length_days,
                    "holidays": 0,
                    "contributors": metrics["contributor_count"],
                }
            )

        return historical_data

    def get_week_issues(self, start_date: str, end_date: str) -> list[dict]:
        """Get issues completed in a date range (for Kanban mode)."""
        jql = f'project = {self.project_key} AND resolved >= "{start_date}" AND resolved <= "{end_date}"'

        output = self._run_jira_command(
            ["issue", "list", "--jql", jql, "--plain", "--columns", "key,status,storypoints,resolved", "--no-headers"]
        )

        issues = []
        for line in output.strip().split("\n"):
            if not line.strip():
                continue
            parts = line.split("\t")
            if len(parts) >= 1:
                story_points_raw = parts[2].strip() if len(parts) > 2 else "0"
                try:
                    story_points = float(story_points_raw) if story_points_raw else 0.0
                except ValueError:
                    story_points = 0.0

                issues.append(
                    {
                        "key": parts[0].strip(),
                        "status": parts[1].strip() if len(parts) > 1 else "",
                        "story_points": story_points,
                        "resolved": parts[3].strip() if len(parts) > 3 else "",
                    }
                )

        return issues

    def get_kanban_metrics(self, weeks: int = 4, start_date: str = None, end_date: str = None) -> list[dict[str, Any]]:
        """
        Get Kanban metrics grouped by week.

        Args:
            weeks: Number of weeks to look back (default: 4)
            start_date: Optional start date (YYYY-MM-DD), overrides weeks param
            end_date: Optional end date (YYYY-MM-DD), defaults to today

        Returns formatted data compatible with sprint analytics structure.
        """
        if not end_date:
            end_dt = date.today()
        else:
            end_dt = datetime.fromisoformat(end_date).date()

        if not start_date:
            start_dt = end_dt - timedelta(days=weeks * 7)
        else:
            start_dt = datetime.fromisoformat(start_date).date()

        issues = self.get_week_issues(start_dt.isoformat(), end_dt.isoformat())

        weeks_data = {}

        for issue in issues:
            resolved_str = issue.get("resolved", "")
            if not resolved_str:
                continue

            try:
                resolved_date = datetime.fromisoformat(resolved_str.split("T")[0])
                week_start = resolved_date - timedelta(days=resolved_date.weekday())
                week_label = f"Week of {week_start.strftime('%b %d')}"

                if week_label not in weeks_data:
                    weeks_data[week_label] = {
                        "sprint": week_label,
                        "committed": 0,
                        "completed": 0,
                        "carryover": 0,
                        "sprint_length_days": 7,
                        "holidays": 0,
                        "contributors": [],
                    }

                weeks_data[week_label]["completed"] += issue["story_points"]
            except Exception:
                continue

        result = sorted(weeks_data.values(), key=lambda x: x["sprint"])
        return result
