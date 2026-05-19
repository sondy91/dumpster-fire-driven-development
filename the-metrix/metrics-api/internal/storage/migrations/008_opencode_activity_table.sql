-- Migration 008: Add opencode_activity table for granular per-project per-day metrics
-- Replaces aggregate opencode_sessions approach with detailed activity tracking

CREATE TABLE IF NOT EXISTS opencode_activity (
    id TEXT PRIMARY KEY,
    date TEXT NOT NULL,              -- YYYY-MM-DD
    developer TEXT NOT NULL,
    project_path TEXT NOT NULL,      -- Full path like /home/user/projects/repo
    project_name TEXT,               -- Extracted repo/folder name (optional)

    -- Action counts
    bash_commands INTEGER DEFAULT 0,
    file_reads INTEGER DEFAULT 0,
    file_edits INTEGER DEFAULT 0,
    file_writes INTEGER DEFAULT 0,
    searches INTEGER DEFAULT 0,

    -- Time saved (in hours)
    time_saved_bash REAL DEFAULT 0.0,
    time_saved_reads REAL DEFAULT 0.0,
    time_saved_edits REAL DEFAULT 0.0,
    time_saved_writes REAL DEFAULT 0.0,
    time_saved_searches REAL DEFAULT 0.0,
    time_saved_total REAL DEFAULT 0.0,

    -- Additional metrics
    avg_complexity_score REAL DEFAULT 0.0,  -- Bash command complexity multiplier
    session_count INTEGER DEFAULT 0,         -- Number of OpenCode sessions that day
    chars_typed INTEGER DEFAULT 0,           -- Total chars for efficiency calculations

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_activity_date ON opencode_activity(date);
CREATE INDEX IF NOT EXISTS idx_activity_developer ON opencode_activity(developer);
CREATE INDEX IF NOT EXISTS idx_activity_project_name ON opencode_activity(project_name);
CREATE INDEX IF NOT EXISTS idx_activity_date_dev ON opencode_activity(date, developer);

-- Unique constraint: one row per developer per project per day
CREATE UNIQUE INDEX IF NOT EXISTS idx_activity_unique ON opencode_activity(date, developer, project_path);
