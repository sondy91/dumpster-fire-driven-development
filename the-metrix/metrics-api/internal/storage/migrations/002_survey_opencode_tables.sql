-- Migration 002: Add survey and OpenCode session tracking tables

-- Survey responses table
CREATE TABLE IF NOT EXISTS survey_responses (
    id TEXT PRIMARY KEY,
    submitted_at TEXT NOT NULL,
    name TEXT DEFAULT 'Anonymous',
    sprint TEXT NOT NULL,
    flow INTEGER CHECK(flow BETWEEN 1 AND 5),
    ai_satisfaction INTEGER CHECK(ai_satisfaction BETWEEN 1 AND 5),
    ai_speed INTEGER CHECK(ai_speed BETWEEN 1 AND 5),
    ai_code_pct INTEGER CHECK(ai_code_pct BETWEEN 0 AND 100),
    devex INTEGER CHECK(devex BETWEEN 1 AND 5),
    blockers TEXT, -- JSON array stored as text
    improvement TEXT,
    ai_tool TEXT,
    space_performance INTEGER CHECK(space_performance BETWEEN 1 AND 5),
    space_collaboration INTEGER CHECK(space_collaboration BETWEEN 1 AND 5),
    space_activity INTEGER CHECK(space_activity BETWEEN 1 AND 5),
    space_efficiency INTEGER CHECK(space_efficiency BETWEEN 1 AND 5),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_survey_sprint ON survey_responses(sprint);
CREATE INDEX IF NOT EXISTS idx_survey_submitted ON survey_responses(submitted_at);

-- OpenCode sessions table
CREATE TABLE IF NOT EXISTS opencode_sessions (
    id TEXT PRIMARY KEY,
    saved_at TEXT NOT NULL,
    developer TEXT NOT NULL,
    project TEXT NOT NULL,
    sprint TEXT NOT NULL,
    bash INTEGER DEFAULT 0,
    reads INTEGER DEFAULT 0,
    edits INTEGER DEFAULT 0,
    searches INTEGER DEFAULT 0,
    time_saved_bash REAL DEFAULT 0.0,
    time_saved_reads REAL DEFAULT 0.0,
    time_saved_edits REAL DEFAULT 0.0,
    time_saved_searches REAL DEFAULT 0.0,
    time_saved_total REAL DEFAULT 0.0,
    session_duration INTEGER DEFAULT 0, -- in minutes
    raw_text TEXT, -- original CLI output
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_opencode_sprint ON opencode_sessions(sprint);
CREATE INDEX IF NOT EXISTS idx_opencode_developer ON opencode_sessions(developer);
CREATE INDEX IF NOT EXISTS idx_opencode_project ON opencode_sessions(project);
CREATE INDEX IF NOT EXISTS idx_opencode_saved ON opencode_sessions(saved_at);
