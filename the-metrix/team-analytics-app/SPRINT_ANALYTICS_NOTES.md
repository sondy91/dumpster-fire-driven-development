# Sprint Analytics - Architecture & Known Issues

## Architecture

### Skeleton Loading Pattern

**Problem:** HTMX doesn't execute `<script>` tags in swapped content by default (security).

**Solution:** Split into two templates:
- `sprint_skeleton.html` - Instant page load with loading skeletons, contains Chart.js CDN and chart initialization logic
- `sprint_content.html` - Fragment loaded via HTMX with HTML + embedded JSON data

**Flow:**
1. User navigates to `/sprint` → Server returns `sprint_skeleton.html` (skeleton=true, default)
2. Skeleton renders instantly with loading placeholders
3. HTMX attribute `hx-get="/sprint?skeleton=false&..."` fetches actual content
4. Content returns HTML + `<script type="application/json" id="chartData">` with chart data
5. `htmx:afterSettle` event fires in skeleton
6. Skeleton script reads `#chartData` JSON element and calls `initializeCharts()`
7. Charts render using Chart.js already loaded in skeleton head

**Key Files:**
- `src/templates/sprint_skeleton.html` - Page shell + chart initialization
- `src/templates/sprint_content.html` - Data fragment (HTML + JSON data element)
- `src/main.py` - Route with `skeleton` parameter (bool, default=true)
- `src/sprint_utils.py` - Data fetching and caching logic

### Caching Strategy

**Server-side cache** (1-hour TTL):
- Cache key: `{project}_{team_id}_{count}_{workflow}_{start_date}_{end_date}`
- Stored in `_SPRINT_CACHE` dict with `(data, timestamp)` tuples
- Logs: `✅ Cache hit!` or `💾 Cached data`

**Demo mode** (`use_cache=true`):
- Uses `HISTORICAL_SPRINTS` from `config.py` (Sprints 10-17)
- Client-side toggle in localStorage: `demo_mode=true/false`
- Synced between nav button and Settings page via `demoModeChanged` custom event

## Known Issues & Data Quality

### Issue 1: Incomplete Sprint Data from Jira

**Symptom:** Real Jira sprints missing `completed`, `committed`, or `carryover` for some sprints.

**Root Cause:** 
- Sprint closed with no issues completed → 0 completed, but committed/carryover also missing
- Team filter (`team_id`) may exclude issues if team field not populated
- Carryover calculation depends on previous sprint data being present

**Example from logs:**
```
Sprint 25.4.24: committed=4, completed=0, carryover=4
Sprint 25.4.22: committed=3, completed=3, carryover=0
Sprint 25.4.19: committed=3, completed=0, carryover=3  # Missing data?
```

**Potential fixes:**
1. Improve carryover calculation to handle gaps in sprint history
2. Add fallback to 0 for missing values instead of omitting datapoints
3. Filter out sprints with no activity (0 across all metrics)
4. Show data quality indicator in UI ("X of Y sprints have complete data")

**Location to fix:** `src/sprint_utils.py` - Functions:
- `get_historical_data_from_jira()`
- `calculate_carryover()` (if exists)

### Issue 2: Contributors Always 1.0 ✅ FIXED

**Symptom:** Performance chart shows contributors=1 for all sprints, even when multiple people worked.

**Root Cause:**
- `JiraClient.get_historical_sprint_data()` line 286 hardcoded: `"contributors": []`
- Empty list → converted to 0 → replaced with 1 as fallback (sprint_utils.py:203)
- Unique assignees were never counted from Jira data

**Fix Applied (2026-04-22):**
- Modified `JiraClient.calculate_sprint_metrics()` to track unique assignees
- Added `contributor_count` to metrics dict: `len(unique_assignees) if unique_assignees else 1`
- Changed `get_historical_sprint_data()` line 286 to use: `"contributors": metrics["contributor_count"]`

**Files Changed:**
- `src/jira_client.py:223-249` - Added unique_assignees set tracking
- `src/jira_client.py:286` - Use calculated contributor_count instead of empty list

**Testing:**
- Clear cache (restart FastAPI or wait 1hr) to see updated data
- Should now see varied contributor counts (1-5) per sprint
- Points/contributor should now reflect team efficiency accurately

### Issue 3: Points Per Contributor Calculation

**Related to Issue 2:** If contributors is always 1, then points/contributor = total points.

**Expected:** Should distribute velocity across team size for efficiency metric.

**Verify calculation:**
```python
points_per_contributor = completed_points / contributors
# If contributors=1 always, this just equals completed_points
```

## Data Quality Checklist

When investigating sprint data issues:

1. **Check raw Jira response:**
   ```bash
   jira sprint list --current --plain
   jira issue list --plain | grep "Sprint 25"
   ```

2. **Verify team filter:**
   ```python
   # In sprint_utils.py, add logging:
   print(f"Team filter: {team_id}")
   print(f"Issues before filter: {len(df)}")
   print(f"Issues after filter: {len(df[df['team']==team_id])}")
   ```

3. **Check carryover logic:**
   - Does it handle first sprint (no previous data)?
   - Does it handle gaps in sprint sequence?
   - Are issues double-counted if moved between sprints?

4. **Validate demo data matches expected patterns:**
   - Demo should show varied contributors (1-5)
   - Demo should show realistic velocity trends
   - Demo should have complete data for all metrics

## Testing Scenarios

### Test 1: Empty Sprint
- **Setup:** Sprint with 0 completed issues
- **Expected:** Charts show 0 completed, committed visible, carryover calculated
- **Actual:** Check if charts render or show errors

### Test 2: Single Contributor
- **Setup:** Sprint where only 1 person worked
- **Expected:** Contributors=1, points/contributor equals velocity
- **Current:** Always shows this (even when shouldn't)

### Test 3: Team Filter
- **Setup:** Change team_id in Settings
- **Expected:** Data changes to show different team's sprints
- **Verify:** Console logs show cache miss and new fetch

### Test 4: Demo Mode Toggle
- **Setup:** Toggle demo mode on/off
- **Expected:** Page reloads, charts show demo vs real data
- **Verify:** Network tab shows use_cache=true/false parameter

## Future Improvements

### Short-term
- [ ] Add data quality indicator ("Using X of Y sprints")
- [ ] Show "No data" message when sprint has 0 activity
- [ ] Add tooltip explaining contributor calculation
- [ ] Debug logging for contributor count per sprint

### Medium-term
- [ ] Support custom date ranges for Kanban workflow
- [ ] Add sprint-over-sprint comparison view
- [ ] Export chart data to CSV
- [ ] Add filters: by assignee, by issue type, by label

### Long-term
- [ ] Real-time updates via WebSocket or polling
- [ ] Predictive velocity forecasting
- [ ] Burndown chart integration
- [ ] Compare multiple teams side-by-side

## Debugging Commands

```bash
# Check FastAPI logs
tail -f /tmp/team-analytics.log

# Test sprint endpoint directly
curl "http://localhost:8000/sprint?skeleton=false&project=XAIP&team_id=2490&count=5"

# Inspect chart data JSON
curl -s "http://localhost:8000/sprint?skeleton=false&use_cache=false" | grep -A 50 'id="chartData"'

# Check Jira CLI data
jira sprint list --current --plain
jira issue list --jql "sprint in openSprints() AND team=2490" --plain

# Clear cache and reload
# (Restart FastAPI or wait 1 hour)
```

## Related Files

- `src/main.py` - `/sprint` route handler
- `src/sprint_utils.py` - Data fetching and calculation logic
- `src/api_client.py` - Jira CLI wrapper (if used)
- `src/config.py` - Demo data (`HISTORICAL_SPRINTS`)
- `src/templates/sprint_skeleton.html` - Page shell + chart init
- `src/templates/sprint_content.html` - Data fragment

---

**Last Updated:** 2026-04-22  
**Author:** OpenCode Agent  
**Related Issues:** XAIP-TBD (Sprint Analytics Data Quality)
