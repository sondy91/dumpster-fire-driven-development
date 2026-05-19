-- Metrics API Database Schema
-- SQLite database for storing user mappings, repo tracking, config, and cached metrics

-- User mapping table (E-number as common identifier across systems)
CREATE TABLE IF NOT EXISTS users (
    e_number TEXT PRIMARY KEY NOT NULL,
    github_username TEXT,
    jira_username TEXT,
    git_email TEXT,
    name TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_github ON users(github_username);
CREATE INDEX IF NOT EXISTS idx_users_jira ON users(jira_username);
CREATE INDEX IF NOT EXISTS idx_users_git_email ON users(git_email);

-- Repository tracking table
CREATE TABLE IF NOT EXISTS repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    org TEXT,
    full_name TEXT UNIQUE NOT NULL,
    is_private BOOLEAN DEFAULT 0,
    clone_path TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_repos_full_name ON repos(full_name);
CREATE INDEX IF NOT EXISTS idx_repos_is_private ON repos(is_private);

-- Configuration key-value storage
CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Metric cache for performance optimization
CREATE TABLE IF NOT EXISTS metric_cache (
    cache_key TEXT PRIMARY KEY NOT NULL,
    data TEXT NOT NULL,  -- JSON data
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_metric_cache_expires ON metric_cache(expires_at);

-- Trigger to update updated_at on users table
CREATE TRIGGER IF NOT EXISTS update_users_timestamp
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE e_number = NEW.e_number;
END;

-- Trigger to update updated_at on repos table
CREATE TRIGGER IF NOT EXISTS update_repos_timestamp
AFTER UPDATE ON repos
FOR EACH ROW
BEGIN
    UPDATE repos SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to update updated_at on config table
CREATE TRIGGER IF NOT EXISTS update_config_timestamp
AFTER UPDATE ON config
FOR EACH ROW
BEGIN
    UPDATE config SET updated_at = CURRENT_TIMESTAMP WHERE key = NEW.key;
END;
