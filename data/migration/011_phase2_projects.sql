CREATE TABLE IF NOT EXISTS projects (
	id TEXT PRIMARY KEY,
	workspace_id TEXT NOT NULL,
	name TEXT NOT NULL,
	slug TEXT NOT NULL,
	summary TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
