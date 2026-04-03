SELECT workspace_mode, COALESCE(sample_initialized_at, ''), COALESCE(live_initialized_at, '')
FROM workspace_settings
WHERE workspace_id = ?;
