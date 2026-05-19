CREATE TABLE IF NOT EXISTS feedback (
    id TEXT PRIMARY KEY,
    timestamp TEXT NOT NULL,
    page TEXT NOT NULL,
    user TEXT,
    user_agent TEXT,
    kind TEXT NOT NULL,
    type TEXT,
    text TEXT,
    component TEXT,
    sentiment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_feedback_page ON feedback(page);
CREATE INDEX IF NOT EXISTS idx_feedback_kind ON feedback(kind);
CREATE INDEX IF NOT EXISTS idx_feedback_created_at ON feedback(created_at DESC);
