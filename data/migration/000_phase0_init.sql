CREATE TABLE IF NOT EXISTS migrations (
	name TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL,
	display_name TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS signals (
	id TEXT PRIMARY KEY,
	source_type TEXT NOT NULL,
	source_id TEXT NOT NULL,
	title TEXT NOT NULL,
	body TEXT NOT NULL,
	url TEXT,
	metadata_json TEXT,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS clusters (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	summary TEXT NOT NULL,
	severity TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cluster_signals (
	cluster_id TEXT NOT NULL,
	signal_id TEXT NOT NULL,
	PRIMARY KEY (cluster_id, signal_id),
	FOREIGN KEY(cluster_id) REFERENCES clusters(id),
	FOREIGN KEY(signal_id) REFERENCES signals(id)
);

CREATE TABLE IF NOT EXISTS work_objects (
	id TEXT PRIMARY KEY,
	cluster_id TEXT,
	title TEXT NOT NULL,
	summary TEXT NOT NULL,
	goal TEXT NOT NULL,
	status TEXT NOT NULL,
	priority TEXT NOT NULL,
	owner_type TEXT NOT NULL,
	owner_id TEXT NOT NULL,
	prototype_url TEXT,
	prototype_note TEXT,
	release_ready_at TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS decisions (
	id TEXT PRIMARY KEY,
	work_object_id TEXT NOT NULL,
	title TEXT NOT NULL,
	summary TEXT NOT NULL,
	rationale TEXT NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY(work_object_id) REFERENCES work_objects(id)
);

CREATE TABLE IF NOT EXISTS artifacts (
	id TEXT PRIMARY KEY,
	work_object_id TEXT NOT NULL,
	artifact_type TEXT NOT NULL,
	title TEXT NOT NULL,
	body TEXT NOT NULL,
	metadata_json TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY(work_object_id) REFERENCES work_objects(id)
);

CREATE TABLE IF NOT EXISTS draft_plans (
	work_object_id TEXT PRIMARY KEY,
	overview TEXT NOT NULL,
	implementation_steps_json TEXT NOT NULL,
	edge_cases_json TEXT NOT NULL,
	open_questions_json TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY(work_object_id) REFERENCES work_objects(id)
);
