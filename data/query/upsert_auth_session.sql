INSERT INTO sessions (
	id,
	user_id,
	created_at,
	refresh_token,
	workos_session_id,
	organization_id,
	organization_name,
	access_token_expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	user_id = excluded.user_id,
	refresh_token = excluded.refresh_token,
	workos_session_id = excluded.workos_session_id,
	organization_id = excluded.organization_id,
	organization_name = excluded.organization_name,
	access_token_expires_at = excluded.access_token_expires_at;
