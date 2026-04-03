CREATE TABLE IF NOT EXISTS workspace_settings (
	workspace_id TEXT PRIMARY KEY,
	workspace_mode TEXT NOT NULL,
	sample_initialized_at TEXT,
	live_initialized_at TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
