INSERT INTO workspace_settings (
	workspace_id,
	workspace_mode,
	sample_initialized_at,
	live_initialized_at,
	created_at,
	updated_at
)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(workspace_id) DO UPDATE SET
	workspace_mode = excluded.workspace_mode,
	sample_initialized_at = COALESCE(workspace_settings.sample_initialized_at, excluded.sample_initialized_at),
	live_initialized_at = COALESCE(workspace_settings.live_initialized_at, excluded.live_initialized_at),
	updated_at = excluded.updated_at;
