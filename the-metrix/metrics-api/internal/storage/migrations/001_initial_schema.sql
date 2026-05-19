-- Initial schema for metrics-api SQLite database
-- Created: 2026-04-16
-- Purpose: Store user mappings, repository config, and cached metrics

-- Users table: Maps e-numbers to Jira/GitHub identities
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    e_number TEXT NOT NULL UNIQUE,
    jira_account_id TEXT,
    github_username TEXT,
    name TEXT,
    email TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_e_number ON users(e_number);
CREATE INDEX IF NOT EXISTS idx_users_jira ON users(jira_account_id);
CREATE INDEX IF NOT EXISTS idx_users_github ON users(github_username);

-- Trigger to update users.updated_at
CREATE TRIGGER IF NOT EXISTS users_updated_at
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Repositories table: Git repositories to track for metrics
CREATE TABLE IF NOT EXISTS repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    repo_type TEXT NOT NULL DEFAULT 'work' CHECK(repo_type IN ('work', 'private')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_repos_type ON repos(repo_type);

-- Trigger to update repos.updated_at
CREATE TRIGGER IF NOT EXISTS repos_updated_at
AFTER UPDATE ON repos
FOR EACH ROW
BEGIN
    UPDATE repos SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Config table: Key-value configuration store
CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Trigger to update config.updated_at
CREATE TRIGGER IF NOT EXISTS config_updated_at
AFTER UPDATE ON config
FOR EACH ROW
BEGIN
    UPDATE config SET updated_at = CURRENT_TIMESTAMP WHERE key = OLD.key;
END;

-- Metric cache table: Store computed metrics with TTL
CREATE TABLE IF NOT EXISTS metric_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_key TEXT NOT NULL UNIQUE,
    metric_type TEXT NOT NULL,
    e_number TEXT,
    start_date DATE,
    end_date DATE,
    data TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cache_key ON metric_cache(cache_key);
CREATE INDEX IF NOT EXISTS idx_cache_expiry ON metric_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_cache_e_number ON metric_cache(e_number);
CREATE INDEX IF NOT EXISTS idx_cache_type ON metric_cache(metric_type);

-- No trigger for metric_cache.updated_at - cache entries are immutable (insert or delete only)
